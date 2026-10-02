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

	requiredGuidelines := []string{
		"PREFIX FLUSHING & STALE LOCKS ARE AUTOMATIC",
		"CLEAN ZERO",
		"SINGLE-VARIABLE ISOLATION",
		"VERIFY EXCEPTION CAUSALITY",
		"COMPOSITOR SANDBOXING OVER ENGINE FLAGS",
		"ESCALATION & DEEP DOCUMENTATION",
	}

	for _, req := range requiredGuidelines {
		found := false
		for _, g := range dump.AIAssistantGuidelines {
			if strings.Contains(g, req) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected %q in ai_assistant_guidelines", req)
		}
	}

	if len(dump.LifecycleHooks.ManagedInternally) == 0 {
		t.Errorf("Expected ManagedInternally to be populated in lifecycle_hooks")
	}

	requiredBestPractices := []string{
		"clean_zero_vs_safe_defaults",
		"clean_zero_baseline",
		"single_variable_delta_isolation",
		"verify_exception_causality",
		"compositor_sandboxing_wayland",
		"wine_prefix_lifecycle_and_stale_locks",
	}

	for _, bp := range requiredBestPractices {
		if _, ok := dump.BestPractices[bp]; !ok {
			t.Errorf("Expected %q in best_practices", bp)
		}
	}

	if _, ok := dump.DocumentationLinks["troubleshooting_methodology"]; !ok {
		t.Errorf("Expected troubleshooting_methodology in documentation_links")
	}
}
