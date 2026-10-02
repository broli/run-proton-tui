package runner

import (
	"path/filepath"
	"strings"
)

// ExeType defines the classification of a Windows binary.
type ExeType string

const (
	ExeType2DUtility ExeType = "2D Utility / Launcher / Installer"
	ExeType3DGame    ExeType = "3D Game Engine"
)

// ClassificationResult provides the detected category and recommended defaults with clear rationale.
type ClassificationResult struct {
	Type               ExeType
	Rationale          string
	RecommendGamescope bool
	RecommendPrimeRun  bool
	RecommendPCores    bool
}

var utilityKeywords = []string{
	"launcher", "update", "setup", "patch", "installer", "unins",
	"repair", "redist", "dxsetup", "vcredist", "crashreport",
	"crashhandler", "config", "service", "helper", "anticheat",
	"ace-", "cef", "webengine", "platformprocess", "games.exe",
}

func is2DUtility(baseName string) bool {
	for _, kw := range utilityKeywords {
		if strings.Contains(baseName, kw) {
			return true
		}
	}
	return false
}

// ClassifyExecutable analyzes the filename and path of an executable to suggest optimal execution settings.
// Crucially, it preserves user agency: these are recommendations that the user can freely toggle.
func ClassifyExecutable(exePath string) ClassificationResult {
	base := strings.ToLower(filepath.Base(exePath))

	if is2DUtility(base) {
		return ClassificationResult{
			Type:               ExeType2DUtility,
			Rationale:          "Detected 2D launcher / setup utility (often Qt5/CEF WebEngine). Recommended: run with system default GPU (no offload) and without Gamescope to prevent NVAPI double-free crashes.",
			RecommendGamescope: false,
			RecommendPrimeRun:  false,
			RecommendPCores:    false,
		}
	}

	// 3D Game binaries
	return ClassificationResult{
		Type:               ExeType3DGame,
		Rationale:          "Detected 3D game executable. Recommended: NVIDIA dGPU offload + Gamescope sandboxing (1080p) + P-Core pinning for peak FPS and stutter prevention.",
		RecommendGamescope: true,
		RecommendPrimeRun:  true,
		RecommendPCores:    true,
	}
}
