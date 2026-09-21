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
// 1. Manages power profile
// 2. Builds Proton environment
// 3. Generates isolated runner script with gamescope-clip-bridge and child process tracking
// 4. Executes with or without Gamescope sandbox
// 5. Cleans up locks and flushes prefix upon exit
// 6. Restores power profile
func RunGame(ctx context.Context, opts LaunchOptions) (*SessionResult, error) {
	cfg := opts.Config
	gameDir := opts.GameDir
	prefixDir := filepath.Join(gameDir, "proton-prefix")
	logsDir, _ := diagnostics.EnsureLogsDir(gameDir)

	// 1. Ensure prefix and pfx subdirectories exist before Proton starts
	// Proton's filelock requires the parent directory (STEAM_COMPAT_DATA_PATH) to exist
	// in order to acquire pfx.lock and initialize wineboot without crashing.
	pfxSubDir := filepath.Join(prefixDir, "pfx")
	if err := os.MkdirAll(pfxSubDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create proton prefix directory: %w", err)
	}

	// 2. Pre-flight health and permissions check (+x on target and *Shipping.exe)
	diagnostics.RunPreflightCheck(gameDir, cfg.TargetExe, prefixDir)

	// 3. Pre-clean abandoned locks and flush any lingering wineserver holding this prefix
	_ = prefix.Flush(prefixDir, cfg.ProtonPath)
	_, _ = prefix.CleanStaleLocks()

	// 4. Power Profile Management
	powerMgr := hardware.NewPowerManager()
	if cfg.ManagePower {
		_ = powerMgr.SetPerformance()
		defer powerMgr.Restore()
	}

	// 5. Build Proton Environment
	extraOverrides := ""
	if len(cfg.DLLOverrides) > 0 {
		var parts []string
		for dll, mode := range cfg.DLLOverrides {
			parts = append(parts, fmt.Sprintf("%s=%s", dll, mode))
		}
		extraOverrides = strings.Join(parts, ";")
	}

	envOpts := proton.EnvOptions{
		PrefixDir:      prefixDir,
		GameDir:        gameDir,
		AppID:          cfg.AppID,
		UseXalia:       cfg.UseXalia,
		EnableLogging:  cfg.EnableLogging,
		ExtraOverrides: extraOverrides,
	}
	envMap := proton.BuildEnvironment(envOpts)
	envSlice := proton.EnvSlice(envMap)


	// 4. Resolve Executable and Directory
	fullExePath := filepath.Join(gameDir, cfg.TargetExe)
	exeDir := filepath.Dir(fullExePath)
	exeName := filepath.Base(fullExePath)

	// 5. Command Prefix (CPU pinning + prime-run)
	cmdPrefix := ""
	if cfg.UsePrimeRun {
		if _, err := exec.LookPath("prime-run"); err == nil {
			cmdPrefix = "prime-run "
		}
	}
	if cfg.UsePCores && cfg.PCoresMask != "" {
		cmdPrefix = fmt.Sprintf("taskset -c %s %s", cfg.PCoresMask, cmdPrefix)
	}

	// 6. Generate Transient Wrapper Script
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

	wrapperContent := fmt.Sprintf(`#!/bin/bash
cd "%s"
echo "$DISPLAY" > /tmp/gamescope-rpt-display
%s

%s"%s" waitforexitandrun ./"%s" "$@"
EXIT_CODE=$?

# Wait for patchers / updaters if spawned
sleep 0.8
while pgrep -u %d -f '(Updater|7zg|Patch)\.exe' >/dev/null 2>&1; do
    sleep 1
done

# Wait for relaunched game if updater restarted it
while pgrep -u %d -f '(Games|Launcher)\.exe' >/dev/null 2>&1; do
    sleep 1
done

if [ -n "$CLIP_BRIDGE_PID" ]; then
    kill "$CLIP_BRIDGE_PID" 2>/dev/null
fi
rm -f /tmp/gamescope-rpt-display 2>/dev/null

exit $EXIT_CODE
`, exeDir, bridgeSnippet, cmdPrefix, cfg.ProtonPath, exeName, uid, uid)

	if err := os.WriteFile(wrapperScript, []byte(wrapperContent), 0755); err != nil {
		return nil, fmt.Errorf("failed to create runner wrapper: %w", err)
	}
	defer func() {
		_ = os.Remove(wrapperScript)
		_ = prefix.Flush(prefixDir, cfg.ProtonPath)
		_, _ = prefix.CleanStaleLocks()
	}()

	// 7. Prepare Process Command (with or without Gamescope)
	var cmd *exec.Cmd
	timestamp := time.Now().Format("2006-01-02_15-04-05")
	sessionLogPath := filepath.Join(logsDir, fmt.Sprintf("game_%s.log", timestamp))

	logFile, _ := os.Create(sessionLogPath)
	defer func() {
		if logFile != nil {
			_ = logFile.Close()
		}
	}()

	combinedArgs := append([]string{wrapperScript}, opts.ExtraArgs...)

	hasGamescope := false
	if _, err := exec.LookPath("gamescope"); err == nil {
		hasGamescope = true
	}

	if cfg.UseGamescope && hasGamescope {
		// Output to preferred monitor (HDMI-A-1 by default)
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
		if cfg.GamescopeOutput != "" {
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
