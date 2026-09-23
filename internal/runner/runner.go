package runner

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/hooks"
	"github.com/broli/run-proton-tui/internal/prefix"
	"github.com/broli/run-proton-tui/internal/proton"
)

// SessionResult contains the outcome of a game launch session.
type SessionResult struct {
	ExitCode      int
	Duration      time.Duration
	LogFile       string
	CrashDetected bool
	AbortedByUser bool
	ErrorOutput   string
}

// LaunchOptions configures the runner execution.
type LaunchOptions struct {
	GameDir    string
	Config     *config.GameConfig
	ExtraArgs  []string
	StdoutPipe io.Writer
	StderrPipe io.Writer
}

// RunGame coordinates the complete execution pipeline:
// 1. Resolves effective profile configuration (2D installer vs 3D game)
// 2. Executes Pre-Launch lifecycle hooks (symlinks, /dev/shm JIT, photo persistence)
// 3. Manages power profile
// 4. Builds Proton environment (with UMU ID and custom environment variables)
// 5. Generates isolated runner script with dynamic process supervision and display export
// 6. Executes with or without Gamescope sandbox
// 7. Cleans up locks, flushes prefix, and executes Post-Exit lifecycle hooks
// 8. Restores power profile
func RunGame(ctx context.Context, opts LaunchOptions) (*SessionResult, error) {
	// 1. Effective Config (incorporates per-binary profiles e.g. Setup.exe vs Game.exe)
	cfg := opts.Config.GetEffectiveConfig(opts.Config.TargetExe)
	gameDir := opts.GameDir
	prefixDir := filepath.Join(gameDir, "proton-prefix")
	logsDir, _ := diagnostics.EnsureLogsDir(gameDir)

	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sessionLogPath := filepath.Join(logsDir, fmt.Sprintf("game_%s.log", timestamp))
	logFile, _ := os.Create(sessionLogPath)
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()

	// 2. Ensure prefix and pfx subdirectories exist before Proton starts
	pfxSubDir := filepath.Join(prefixDir, "pfx")
	if err := os.MkdirAll(pfxSubDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create proton prefix directory: %w", err)
	}

	// 3. Pre-flight health and permissions check (+x on target and *Shipping.exe)
	diagnostics.RunPreflightCheck(gameDir, cfg.TargetExe, prefixDir)

	// 4. Pre-flight inter-game conflict detection
	if conflicts, err := prefix.ListOtherWineProcesses(prefixDir); err == nil && len(conflicts) > 0 {
		if logFile != nil {
			_, _ = fmt.Fprintf(logFile, "[PREFLIGHT:WARN] Detected %d active Wine/Proton process(es) in other prefixes:\n", len(conflicts))
			for _, c := range conflicts {
				_, _ = fmt.Fprintf(logFile, "  - PID %d: %s (Prefix: %s)\n", c.PID, c.ExeName, c.Prefix)
			}
		}
	}

	// 5. Declarative filesystem directives (ensure_dirs, symlinks)
	processFilesystemDirectives(gameDir, prefixDir, cfg.Filesystem, logFile)

	// 6. Pre-clean abandoned locks and flush any lingering wineserver holding this prefix
	_ = prefix.Flush(prefixDir, cfg.ProtonPath)
	_, _ = prefix.CleanStaleLocks()

	// 7. Pre-Launch Lifecycle Hook Execution
	hookOpts := hooks.HookOptions{
		GameDir:        gameDir,
		PrefixDir:      prefixDir,
		TargetExe:      cfg.TargetExe,
		ProtonPath:     cfg.ProtonPath,
		ConfigHookPath: cfg.PreLaunchHook,
		ExtraHookDirs:  cfg.HookDirs,
		LogWriter:      logFile,
	}
	_, _ = hooks.ExecuteHook(ctx, hooks.PreLaunch, hookOpts)

	// 8. Power Profile Management
	powerMgr := hardware.NewPowerManager()
	if cfg.ManagePower {
		_ = powerMgr.SetPerformance()
		defer powerMgr.Restore()
	}

	// 9. Build Proton Environment (incorporates UMU ID, 2D/3D isolation, and custom variables)
	extraOverrides := ""
	if len(cfg.DLLOverrides) > 0 {
		var parts []string
		for dll, mode := range cfg.DLLOverrides {
			parts = append(parts, fmt.Sprintf("%s=%s", dll, mode))
		}
		extraOverrides = strings.Join(parts, ";")
	}

	cls := ClassifyExecutable(cfg.TargetExe)
	is2D := cls.Type == ExeType2DUtility

	envOpts := proton.EnvOptions{
		PrefixDir:      prefixDir,
		GameDir:        gameDir,
		AppID:          cfg.AppID,
		UsePrimeRun:    cfg.UsePrimeRun,
		Is2DUtility:    is2D,
		UseXalia:       cfg.UseXalia,
		EnableLogging:  cfg.EnableLogging,
		ExtraOverrides: extraOverrides,
		UmuID:          cfg.UmuID,
		CustomEnv:      cfg.EnvVars,
	}
	envMap := proton.BuildEnvironment(envOpts)
	envSlice := proton.EnvSlice(envMap)

	// 10. Resolve Executable and Directory
	fullExePath := filepath.Join(gameDir, cfg.TargetExe)
	exeDir := filepath.Dir(fullExePath)
	exeName := filepath.Base(fullExePath)

	// 11. Command Prefix (CPU pinning + prime-run)
	cmdPrefix := ""
	if cfg.UsePrimeRun {
		if _, err := exec.LookPath("prime-run"); err == nil {
			cmdPrefix = "prime-run "
		}
	}
	if cfg.UsePCores && cfg.PCoresMask != "" {
		cmdPrefix = fmt.Sprintf("taskset -c %s %s", cfg.PCoresMask, cmdPrefix)
	}

	// 10. Generate Transient Wrapper Script
	wrapperScript := filepath.Join(gameDir, ".rpt_runner.sh")
	uid := os.Getuid()

	// Locate clipboard bridge if available
	home, _ := os.UserHomeDir()
	bridgeBin := filepath.Join(home, "bin", "gamescope-clip-bridge")
	bridgeSnippet := ""
	if _, err := os.Stat(bridgeBin); err == nil {
		bridgeSnippet = fmt.Sprintf(`
CLIP_BRIDGE_PID=""
if [ -x "%s" ] && [ "$DISPLAY" != ":0" ]; then
    "%s" &
    CLIP_BRIDGE_PID=$!
fi
`, bridgeBin, bridgeBin)
	}

	// Configure dynamic process supervision for updaters / child installers
	waitSnippet := ""
	if len(cfg.WaitProcesses) > 0 {
		waitSnippet = "\n# Supervise configured background/child processes\nsleep 0.8\n"
		for _, proc := range cfg.WaitProcesses {
			waitSnippet += fmt.Sprintf("while pgrep -u %d -f '%s' >/dev/null 2>&1; do sleep 1; done\n", uid, proc)
		}
	}

	displayExport := `echo "$DISPLAY" > /tmp/gamescope-rpt-display`
	displayCleanup := `rm -f /tmp/gamescope-rpt-display 2>/dev/null`
	if cfg.DisplayFile != "" {
		displayExport += fmt.Sprintf("\necho \"$DISPLAY\" > %s", cfg.DisplayFile)
		displayCleanup += fmt.Sprintf("\nrm -f %s 2>/dev/null", cfg.DisplayFile)
	}

	wrapperContent := fmt.Sprintf(`#!/bin/bash
cd "%s"
%s
%s

%s"%s" waitforexitandrun ./"%s" "$@"
EXIT_CODE=$?
%s
if [ -n "$CLIP_BRIDGE_PID" ]; then
    kill "$CLIP_BRIDGE_PID" 2>/dev/null
fi
%s

exit $EXIT_CODE
`, exeDir, displayExport, bridgeSnippet, cmdPrefix, cfg.ProtonPath, exeName, waitSnippet, displayCleanup)

	if err := os.WriteFile(wrapperScript, []byte(wrapperContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to create runner wrapper: %w", err)
	}
	defer func() {
		_ = os.Remove(wrapperScript)
		_ = prefix.Flush(prefixDir, cfg.ProtonPath)
		_, _ = prefix.CleanStaleLocks()

		// Post-Exit Lifecycle Hook Execution
		hookOpts.ConfigHookPath = cfg.PostExitHook
		hookOpts.ExtraHookDirs = cfg.HookDirs
		_, _ = hooks.ExecuteHook(context.Background(), hooks.PostExit, hookOpts)
	}()

	// 12. Prepare Process Command (with or without Gamescope)
	var cmd *exec.Cmd
	combinedArgs := append([]string{wrapperScript}, opts.ExtraArgs...)

	hasGamescope := false
	if _, err := exec.LookPath("gamescope"); err == nil {
		hasGamescope = true
	}

	if cfg.UseGamescope && hasGamescope {
		// Output to preferred monitor (HDMI-A-1 by default, or auto)
		// Gamescope runs on host iGPU (no prime-run on gamescope itself to prevent KWin crash!)
		gsArgs := []string{
			"-W", strconv.Itoa(cfg.GamescopeWidth),
			"-H", strconv.Itoa(cfg.GamescopeHeight),
			"-w", strconv.Itoa(cfg.GamescopeWidth),
			"-h", strconv.Itoa(cfg.GamescopeHeight),
			"-r", strconv.Itoa(cfg.GamescopeRefresh),
			"--force-windows-fullscreen",
			"-f",
		}
		if cfg.GamescopeOutput != "" && !strings.EqualFold(cfg.GamescopeOutput, "auto") {
			gsArgs = append(gsArgs, "--prefer-output", cfg.GamescopeOutput)
		}
		gsArgs = append(gsArgs, "--", "/bin/bash")
		gsArgs = append(gsArgs, combinedArgs...)

		cmd = exec.CommandContext(ctx, "gamescope", gsArgs...)
	} else {
		cmd = exec.CommandContext(ctx, "/bin/bash", combinedArgs...)
	}

	cmd.Dir = gameDir
	cmd.Env = envSlice

	// Place child processes into their own process group so signals can be propagated
	// cleanly to gamescope, proton, and wineserver without leaking dangling wine processes.
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	cmd.Cancel = func() error {
		if cmd.Process != nil && cmd.Process.Pid > 0 {
			return syscall.Kill(-cmd.Process.Pid, syscall.SIGTERM)
		}
		return nil
	}
	cmd.WaitDelay = 3 * time.Second

	// Pipe output
	if logFile != nil {
		if opts.StdoutPipe != nil {
			cmd.Stdout = io.MultiWriter(opts.StdoutPipe, logFile)
		} else {
			cmd.Stdout = logFile
		}
		if opts.StderrPipe != nil {
			cmd.Stderr = io.MultiWriter(opts.StderrPipe, logFile)
		} else {
			cmd.Stderr = logFile
		}
	}

	startTime := time.Now()
	err := cmd.Run()
	duration := time.Since(startTime)

	result := &SessionResult{
		Duration: duration,
		LogFile:  sessionLogPath,
	}

	if ctx.Err() != nil {
		result.AbortedByUser = true
		result.ExitCode = 130
	} else if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = 1
		}
	}

	// Crash Detection: exited with error or died within 15 seconds
	if !result.AbortedByUser && (result.ExitCode != 0 || (duration < 15*time.Second && result.ExitCode != 0)) {
		result.CrashDetected = true
	}

	return result, nil
}

// processFilesystemDirectives creates required directories and declarative symlinks.
func processFilesystemDirectives(gameDir, prefixDir string, fs config.FilesystemConfig, logWriter io.Writer) {
	home, _ := os.UserHomeDir()
	resolvePath := func(p string) string {
		p = strings.ReplaceAll(p, "{PREFIX}", prefixDir)
		p = strings.ReplaceAll(p, "{GAME_DIR}", gameDir)
		p = strings.ReplaceAll(p, "{HOST_HOME}", home)
		p = strings.ReplaceAll(p, "{HOST_PICTURES}", filepath.Join(home, "Pictures"))
		if strings.HasPrefix(p, "~/") {
			p = filepath.Join(home, p[2:])
		}
		return p
	}

	for _, dir := range fs.EnsureDirs {
		target := resolvePath(dir)
		if target != "" {
			_ = os.MkdirAll(target, 0755)
		}
	}

	for _, sym := range fs.Symlinks {
		src := resolvePath(sym.Source)
		dst := resolvePath(sym.Target)
		if src == "" || dst == "" {
			continue
		}
		_ = os.MkdirAll(filepath.Dir(dst), 0755)
		_ = os.MkdirAll(src, 0755)
		if fi, err := os.Lstat(dst); err == nil {
			if fi.Mode()&os.ModeSymlink != 0 {
				_ = os.Remove(dst)
			} else if fi.IsDir() {
				entries, _ := os.ReadDir(dst)
				if len(entries) == 0 {
					_ = os.Remove(dst)
				}
			}
		}
		_ = os.Symlink(src, dst)
	}
}
