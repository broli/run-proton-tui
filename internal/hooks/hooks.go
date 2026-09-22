package hooks

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

// HookType identifies the lifecycle phase of a hook script.
type HookType string

const (
	PreLaunch HookType = "pre_launch"
	PostExit  HookType = "post_exit"
)

// HookOptions provides the context and parameters for hook execution.
type HookOptions struct {
	GameDir          string
	PrefixDir        string
	TargetExe        string
	ProtonPath       string
	GamescopeDisplay string
	ConfigHookPath   string
	LogWriter        io.Writer
}

// HookResult contains execution details of a lifecycle hook.
type HookResult struct {
	Executed   bool
	ScriptPath string
	ExitCode   int
	Output     string
}

// ResolveHook cascades through potential hook locations in order:
// 1. Explicitly configured path from .proton-config.toml
// 2. Local game directory ($PWD/.rpt/hooks/<type>.sh)
// 3. User config directory ($HOME/.config/rpt/hooks/<type>.sh)
// 4. User legacy directory ($HOME/.rpt/hooks/<type>.sh)
func ResolveHook(gameDir string, hookType HookType, configHookPath string) string {
	scriptName := string(hookType) + ".sh"

	// 1. Configured hook path
	if configHookPath != "" {
		target := configHookPath
		if !filepath.IsAbs(target) {
			target = filepath.Join(gameDir, target)
		}
		if fi, err := os.Stat(target); err == nil && !fi.IsDir() {
			return target
		}
	}

	// 2. Local game directory: $PWD/.rpt/hooks/<name>.sh
	localHook := filepath.Join(gameDir, ".rpt", "hooks", scriptName)
	if fi, err := os.Stat(localHook); err == nil && !fi.IsDir() {
		return localHook
	}

	// 3. User home directory
	if home, err := os.UserHomeDir(); err == nil {
		userConfigHook := filepath.Join(home, ".config", "rpt", "hooks", scriptName)
		if fi, err := os.Stat(userConfigHook); err == nil && !fi.IsDir() {
			return userConfigHook
		}

		userLegacyHook := filepath.Join(home, ".rpt", "hooks", scriptName)
		if fi, err := os.Stat(userLegacyHook); err == nil && !fi.IsDir() {
			return userLegacyHook
		}
	}

	return ""
}

// ExecuteHook runs the resolved lifecycle hook script with standardized environment variables.
func ExecuteHook(ctx context.Context, hookType HookType, opts HookOptions) (*HookResult, error) {
	scriptPath := ResolveHook(opts.GameDir, hookType, opts.ConfigHookPath)
	if scriptPath == "" {
		return &HookResult{Executed: false}, nil
	}

	// Ensure execution permission
	_ = os.Chmod(scriptPath, 0755)

	cmd := exec.CommandContext(ctx, "/bin/bash", scriptPath)
	cmd.Dir = opts.GameDir

	// Inherit environment and inject standard RPT lifecycle variables
	env := os.Environ()
	env = append(env,
		fmt.Sprintf("RPT_GAME_DIR=%s", opts.GameDir),
		fmt.Sprintf("RPT_PREFIX_DIR=%s", opts.PrefixDir),
		fmt.Sprintf("RPT_TARGET_EXE=%s", opts.TargetExe),
		fmt.Sprintf("RPT_PROTON_PATH=%s", opts.ProtonPath),
		fmt.Sprintf("RPT_GAMESCOPE_DISPLAY=%s", opts.GamescopeDisplay),
		fmt.Sprintf("RPT_HOOK_TYPE=%s", string(hookType)),
	)
	cmd.Env = env

	var outBuf bytes.Buffer
	var writer io.Writer = &outBuf

	if opts.LogWriter != nil {
		writer = io.MultiWriter(&outBuf, opts.LogWriter)
	}

	cmd.Stdout = writer
	cmd.Stderr = writer

	if opts.LogWriter != nil {
		_, _ = fmt.Fprintf(opts.LogWriter, "[HOOK:%s] Starting execution: %s\n", hookType, scriptPath)
	}

	err := cmd.Run()

	res := &HookResult{
		Executed:   true,
		ScriptPath: scriptPath,
		Output:     outBuf.String(),
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			res.ExitCode = exitErr.ExitCode()
		} else {
			res.ExitCode = 1
		}
		if opts.LogWriter != nil {
			_, _ = fmt.Fprintf(opts.LogWriter, "[HOOK:%s] Finished with error (code %d)\n", hookType, res.ExitCode)
		}
		return res, fmt.Errorf("hook %s exited with code %d: %w", hookType, res.ExitCode, err)
	}

	if opts.LogWriter != nil {
		_, _ = fmt.Fprintf(opts.LogWriter, "[HOOK:%s] Finished successfully (code 0)\n", hookType)
	}

	return res, nil
}

// StreamHookOutput prefixes hook output for clean unified logging.
func StreamHookOutput(r io.Reader, w io.Writer, hookType HookType) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		_, _ = fmt.Fprintf(w, "[HOOK:%s] %s\n", hookType, scanner.Text())
	}
}
