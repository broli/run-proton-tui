package quirks

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/config"
)

func TestSubmitQuirk_MITRequired(t *testing.T) {
	cfg := config.NewDefaultConfig()
	_, err := SubmitQuirk("/tmp/MyGame", cfg, false)
	if err == nil {
		t.Fatalf("Expected error when acceptedMIT is false, got nil")
	}
	if !strings.Contains(err.Error(), "MIT License") {
		t.Errorf("Expected MIT License rejection message, got %v", err)
	}
}

func TestGenerateQuirkRecipe(t *testing.T) {
	cfg := config.NewDefaultConfig()
	cfg.AppID = "123456"
	cfg.TargetExe = "Game.exe"
	cfg.UseGamescope = true
	cfg.DLLOverrides["dxgi"] = "n,b"

	recipe, slug := GenerateQuirkRecipe("/tmp/Super Cool Game", cfg)

	if slug != "super-cool-game" {
		t.Errorf("Expected slug 'super-cool-game', got %s", slug)
	}
	if !strings.Contains(recipe, "Super Cool Game") {
		t.Errorf("Recipe missing game title: %s", recipe)
	}
	if !strings.Contains(recipe, "dxgi = \"n,b\"") {
		t.Errorf("Recipe missing DLL override: %s", recipe)
	}
	if !strings.Contains(recipe, "MIT License") {
		t.Errorf("Recipe missing MIT license acknowledgement: %s", recipe)
	}
}
