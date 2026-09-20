package runner

import (
	"testing"
)

func TestClassifyExecutable(t *testing.T) {
	tests := []struct {
		exe       string
		want2D    bool
		wantGames bool
	}{
		{"Launcher.exe", true, false},
		{"Games.exe", true, false},
		{"Setup.exe", true, false},
		{"unins000.exe", true, false},
		{"Endfield.exe", false, true},
		{"Game-Win64-Shipping.exe", false, true},
		{"EldenRing.exe", false, true},
	}

	for _, tt := range tests {
		res := ClassifyExecutable(tt.exe)
		if tt.want2D && res.Type != ExeType2DUtility {
			t.Errorf("ClassifyExecutable(%q) got %v, want 2D utility", tt.exe, res.Type)
		}
		if !tt.want2D && res.Type != ExeType3DGame {
			t.Errorf("ClassifyExecutable(%q) got %v, want 3D game", tt.exe, res.Type)
		}
		if res.RecommendGamescope != tt.wantGames {
			t.Errorf("ClassifyExecutable(%q) RecommendGamescope = %v, want %v", tt.exe, res.RecommendGamescope, tt.wantGames)
		}
	}
}
