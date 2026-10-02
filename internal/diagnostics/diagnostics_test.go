package diagnostics

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunPreflightCheck(t *testing.T) {
	tmpDir := t.TempDir()
	exePath := filepath.Join(tmpDir, "Game.exe")
	// Create executable without +x bit
	_ = os.WriteFile(exePath, []byte("fake-binary"), 0644)

	prefixDir := filepath.Join(tmpDir, "proton-prefix")
	report := RunPreflightCheck(tmpDir, "Game.exe", prefixDir)

	if !report.TargetExeExists {
		t.Errorf("Expected TargetExeExists=true")
	}

	if !report.TargetExeFixed {
		t.Errorf("Expected TargetExeFixed=true (auto-chmod +x)")
	}

	// Verify file is now executable
	info, err := os.Stat(exePath)
	if err != nil {
		t.Fatalf("Failed to stat exe: %v", err)
	}
	if info.Mode()&0111 == 0 {
		t.Errorf("Expected executable bit to be set")
	}
}

func TestGenerateSpecDump(t *testing.T) {
	data, err := GenerateSpecDump("2.1.0")
	if err != nil {
		t.Fatalf("GenerateSpecDump failed: %v", err)
	}

	str := string(data)
	if len(str) == 0 {
		t.Fatalf("Expected non-empty output")
	}
	if !strings.Contains(str, "rpt-spec-v2") {
		t.Errorf("Expected rpt-spec-v2 in output")
	}
	if !strings.Contains(str, "ai_assistant_guidelines") {
		t.Errorf("Expected ai_assistant_guidelines in output")
	}
	if !strings.Contains(str, "hook_recipes") {
		t.Errorf("Expected hook_recipes in output")
	}
	if !strings.Contains(str, "connected_display_outputs") {
		t.Errorf("Expected connected_display_outputs in output")
	}
	if !strings.Contains(str, "lifecycle_hooks") {
		t.Errorf("Expected lifecycle_hooks in output")
	}
	if !strings.Contains(str, "managed_internally") {
		t.Errorf("Expected managed_internally in output")
	}

	var dump SpecDump
	if err := json.Unmarshal(data, &dump); err != nil {
		t.Fatalf("Failed to unmarshal SpecDump JSON: %v", err)
	}

	foundPrefixGuideline := false
	for _, g := range dump.AIAssistantGuidelines {
		if strings.Contains(g, "PREFIX FLUSHING & STALE LOCKS ARE AUTOMATIC") {
			foundPrefixGuideline = true
			break
		}
	}
	if !foundPrefixGuideline {
		t.Errorf("Expected PREFIX FLUSHING & STALE LOCKS ARE AUTOMATIC in ai_assistant_guidelines")
	}

	if len(dump.LifecycleHooks.ManagedInternally) == 0 {
		t.Errorf("Expected ManagedInternally to be populated in lifecycle_hooks")
	}

	if _, ok := dump.BestPractices["wine_prefix_lifecycle_and_stale_locks"]; !ok {
		t.Errorf("Expected wine_prefix_lifecycle_and_stale_locks in best_practices")
	}
}
