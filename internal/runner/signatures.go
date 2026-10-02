package runner

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/pathutil"
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
			Category: "32-bit D3D9 Swapchain Deadlock",
			Keywords: []string{"d3d9: present failed", "d3derr_devicelost", "swapchain creation failed"},
			Observation: "Vintage 32-bit DirectX 9 engine encountered a swapchain presentation deadlock.",
			Recommendation: "If running older 32-bit D3D9 titles, testing with Gamescope disabled ('rpt --gamescope false --now') may help isolate whether nested Xwayland presentation is stalling.",
		},
		{
			Category: "Crash Handler / Fatal Exception Trap",
			Keywords: []string{"crashsight", "crashsightlog", "suspendthread loop failed"},
			Observation: "An internal crash handler caught an unhandled fatal exception or driver abort during startup.",
			Recommendation: "Inspect runtime logs and crash dumps in .logs/ for detailed stack traces, or ask an AI assistant to analyze the logs.",
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

	if home := pathutil.UserHomeDir(); home != "" {
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
