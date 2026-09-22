package hooks

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestResolveHookCascading(t *testing.T) {
	tmpDir := t.TempDir()

	// Initially no hook
	if path := ResolveHook(tmpDir, PreLaunch, ""); path != "" {
		t.Errorf("Expected empty path when no hook exists, got %s", path)
	}

	// Create local .rpt/hooks/pre_launch.sh
	localHooksDir := filepath.Join(tmpDir, ".rpt", "hooks")
	_ = os.MkdirAll(localHooksDir, 0755)
	localScript := filepath.Join(localHooksDir, "pre_launch.sh")
	_ = os.WriteFile(localScript, []byte("#!/bin/bash\necho local\n"), 0755)

	if path := ResolveHook(tmpDir, PreLaunch, ""); path != localScript {
		t.Errorf("Expected local script %s, got %s", localScript, path)
	}

	// Configured hook path takes priority
	customScript := filepath.Join(tmpDir, "custom_pre.sh")
	_ = os.WriteFile(customScript, []byte("#!/bin/bash\necho custom\n"), 0755)

	if path := ResolveHook(tmpDir, PreLaunch, "custom_pre.sh"); path != customScript {
		t.Errorf("Expected custom script %s, got %s", customScript, path)
	}
}

func TestExecuteHook(t *testing.T) {
	tmpDir := t.TempDir()
	scriptContent := `#!/bin/bash
echo "RUNNING_HOOK: dir=$RPT_GAME_DIR target=$RPT_TARGET_EXE"
`
	scriptPath := filepath.Join(tmpDir, "test_hook.sh")
	_ = os.WriteFile(scriptPath, []byte(scriptContent), 0755)

	var logBuf bytes.Buffer
	opts := HookOptions{
		GameDir:          tmpDir,
		PrefixDir:        filepath.Join(tmpDir, "proton-prefix"),
		TargetExe:        "Endfield.exe",
		ProtonPath:       "/mock/proton",
		GamescopeDisplay: ":1",
		ConfigHookPath:   scriptPath,
		LogWriter:        &logBuf,
	}

	res, err := ExecuteHook(context.Background(), PreLaunch, opts)
	if err != nil {
		t.Fatalf("ExecuteHook failed: %v", err)
	}
	if !res.Executed {
		t.Errorf("Expected hook to be executed")
	}
	if res.ExitCode != 0 {
		t.Errorf("Expected exit code 0, got %d", res.ExitCode)
	}
	if !strings.Contains(res.Output, "RUNNING_HOOK: dir="+tmpDir) {
		t.Errorf("Output missing expected environment vars: %s", res.Output)
	}
	if !strings.Contains(logBuf.String(), "[HOOK:pre_launch]") {
		t.Errorf("Log output missing hook prefix: %s", logBuf.String())
	}
}
