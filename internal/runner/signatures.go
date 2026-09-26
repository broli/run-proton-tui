package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

// SignatureDefinition represents an error or stall pattern to detect in session logs.
type SignatureDefinition struct {
	Category       string   `toml:"category"`
	Keywords       []string `toml:"keywords"`
	Observation    string   `toml:"observation"`
	Recommendation string   `toml:"recommendation"`
}

// SignaturesFile defines the schema of an external signatures.toml dictionary.
type SignaturesFile struct {
	Signatures []SignatureDefinition `toml:"signatures"`
}

// GetDefaultSignatures returns the built-in modular signatures dictionary.
func GetDefaultSignatures() []SignatureDefinition {
	return []SignatureDefinition{
		{
			Category: "Anti-Cheat Driver / Kernel Stub Abort",
			Keywords: []string{"ace-base.sys", "status_wine_stub", "probeforwrite", "winedevice.exe", "0x80000100"},
			Observation: "Anti-cheat kernel driver or unsupported kernel stub aborted during initialization.",
			Recommendation: "Check if the game requires a Wine kernel patch (e.g. ntoskrnl.exe ProbeForWrite) or specialized runner. Use a hooks/pre_launch.sh script to verify or apply required patches.",
		},
		{
			Category: "Missing Kernel Stub / Entrypoint",
			Keywords: []string{"status_entrypoint_not_found"},
			Observation: "Wine encountered a missing DLL entrypoint or kernel stub.",
			Recommendation: "Try switching to GE-Proton or Proton Experimental, or configure DLL overrides in .proton-config.toml.",
		},
		{
			Category: "Bink Video / Intro Splash Stall",
			Keywords: []string{"coda transfers", "binkw32"},
			Observation: "Engine stalled while allocating buffers for intro video decompression or 2D splash presentation.",
			Recommendation: "Try toggling Gamescope off ('rpt --gamescope false --now') to allow native Xwayland presentation.",
		},
		{
			Category: "Gamescope & 32-bit Swapchain Stall",
			Keywords: []string{"failed to read wayland events", "vk_wsi_force_swapchain_to_current_extent"},
			Observation: "32-bit DirectX engine or nested compositor swapchain deadlock detected.",
			Recommendation: "Toggle Gamescope OFF ('rpt --gamescope false --now') to run natively on host Wayland / Xwayland.",
		},
	}
}

// LoadSignatures loads signatures combining built-in defaults with external signatures.toml if present.
func LoadSignatures(gameDir string) []SignatureDefinition {
	sigs := GetDefaultSignatures()

	candidates := []string{
		filepath.Join(gameDir, ".rpt", "signatures.toml"),
		filepath.Join(gameDir, "signatures.toml"),
	}

	if home, err := os.UserHomeDir(); err == nil {
		candidates = append(candidates, filepath.Join(home, ".config", "rpt", "signatures.toml"))
	}

	for _, path := range candidates {
		data, err := os.ReadFile(path)
		if err != nil {
			continue
		}

		var file SignaturesFile
		if err := toml.Unmarshal(data, &file); err == nil && len(file.Signatures) > 0 {
			sigs = append(sigs, file.Signatures...)
			break
		}
	}

	return sigs
}

// MatchSignature checks if a log line matches any signature keywords (case-insensitive).
func MatchSignature(line string, sig SignatureDefinition) bool {
	lower := strings.ToLower(line)
	for _, kw := range sig.Keywords {
		if strings.Contains(lower, strings.ToLower(kw)) {
			return true
		}
	}
	return false
}
