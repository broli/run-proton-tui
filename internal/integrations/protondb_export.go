package integrations

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/runner"
)

// ExportOptions holds optional parameters for synthesizing a ProtonDB community report.
type ExportOptions struct {
	AppID      string
	GameTitle  string
	GameDir    string
	Config     *config.GameConfig
	LastResult *runner.SessionResult
	History    []runner.SessionRecord
}

// DetermineRecommendedTier calculates the recommended ProtonDB tier based on session outcomes and tweaks.
func DetermineRecommendedTier(cfg *config.GameConfig, lastResult *runner.SessionResult, history []runner.SessionRecord) string {
	// 1. If the last run crashed or exited with anti-cheat/segfault codes
	if lastResult != nil && (lastResult.CrashDetected || (lastResult.ExitCode != 0 && !lastResult.AbortedByUser)) {
		return "Borked"
	}

	// 2. Check if significant tweaks or quirks were required
	hasTweaks := false
	if cfg != nil {
		if cfg.UseGamescope || cfg.UsePrimeRun || cfg.UsePCores || cfg.UseXalia || len(cfg.DLLOverrides) > 0 || len(cfg.EnvVars) > 0 || cfg.PreLaunchHook != "" {
			hasTweaks = true
		}
		// If custom GE runner or non-default runner is used
		baseRunner := strings.ToLower(filepath.Base(cfg.ProtonPath))
		if strings.Contains(baseRunner, "ge") || strings.Contains(baseRunner, "custom") || strings.Contains(baseRunner, "cachyos") {
			hasTweaks = true
		}
	}

	// 3. Inspect historical sessions if available
	totalRuns := len(history)
	crashCount := 0
	for _, rec := range history {
		if rec.CrashDetected || (rec.ExitCode != 0 && !rec.AbortedByUser) {
			crashCount++
		}
	}

	if totalRuns > 0 && float64(crashCount)/float64(totalRuns) > 0.4 {
		return "Bronze"
	}
	if totalRuns > 0 && crashCount > 0 {
		return "Silver"
	}

	if hasTweaks {
		return "Gold"
	}

	return "Platinum"
}

// detectDistroAndKernel reads /etc/os-release and kernel info.
func detectDistroAndKernel() (string, string) {
	distro := "Linux"
	if f, err := os.Open("/etc/os-release"); err == nil {
		defer f.Close()
		scanner := bufio.NewScanner(f)
		for scanner.Scan() {
			line := scanner.Text()
			if strings.HasPrefix(line, "PRETTY_NAME=") {
				distro = strings.Trim(strings.TrimPrefix(line, "PRETTY_NAME="), "\"")
				break
			}
		}
		_ = scanner.Err()
	}

	kernel := "Unknown"
	if kbytes, err := os.ReadFile("/proc/version"); err == nil {
		fields := strings.Fields(string(kbytes))
		if len(fields) >= 3 {
			kernel = fields[2]
		}
	}

	return distro, kernel
}

// GenerateProtonDBReport constructs a structured Markdown report formatted for ProtonDB.
func GenerateProtonDBReport(opts ExportOptions) string {
	cfg := opts.Config
	if cfg == nil {
		cfg = config.NewDefaultConfig()
	}

	distro, kernel := detectDistroAndKernel()

	// Hardware detection
	gpuInfo, _ := hardware.DetectGPU()
	cpuTopo, _ := hardware.DetectCPUTopology()

	var gpuDesc strings.Builder
	if gpuInfo != nil {
		var vendors []string
		if gpuInfo.HasNvidia {
			vendors = append(vendors, "NVIDIA (dGPU)")
		}
		if gpuInfo.HasAMD {
			vendors = append(vendors, "AMD")
		}
		if gpuInfo.HasIntel {
			vendors = append(vendors, "Intel")
		}
		if len(vendors) > 0 {
			gpuDesc.WriteString(strings.Join(vendors, " / "))
		} else {
			gpuDesc.WriteString("Vulkan-compatible GPU")
		}
		if gpuInfo.HasPrimeRun && cfg.UsePrimeRun {
			gpuDesc.WriteString(" [prime-run offload]")
		}
	} else {
		gpuDesc.WriteString("Linux DRM GPU")
	}

	cpuDesc := "x86_64 Processor"
	if cpuTopo != nil && cpuTopo.ModelName != "" {
		cpuDesc = fmt.Sprintf("%s (%d threads", cpuTopo.ModelName, cpuTopo.TotalThreads)
		if cpuTopo.IsHybrid && cfg.UsePCores {
			cpuDesc += fmt.Sprintf(", P-cores %s pinned)", cfg.PCoresMask)
		} else {
			cpuDesc += ")"
		}
	}

	sessType := os.Getenv("XDG_SESSION_TYPE")
	if sessType == "" {
		sessType = "Wayland/X11"
	}
	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if desktop == "" {
		desktop = "Desktop"
	}

	runnerName := filepath.Base(cfg.ProtonPath)
	if runnerName == "" || runnerName == "." {
		runnerName = "Default Proton"
	}

	tier := DetermineRecommendedTier(cfg, opts.LastResult, opts.History)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("### ProtonDB Report for %s (AppID: %s)\n\n", opts.GameTitle, opts.AppID))

	sb.WriteString("#### 🖥️ System Specs\n")
	sb.WriteString(fmt.Sprintf("- **OS / Kernel**: %s (Kernel %s)\n", distro, kernel))
	sb.WriteString(fmt.Sprintf("- **Session**: %s (%s)\n", sessType, desktop))
	sb.WriteString(fmt.Sprintf("- **CPU**: %s\n", cpuDesc))
	sb.WriteString(fmt.Sprintf("- **GPU**: %s\n", gpuDesc.String()))
	sb.WriteString(fmt.Sprintf("- **Runner**: %s\n\n", runnerName))

	sb.WriteString("#### ⚙️ Configuration & Tweaks Applied\n")
	if cfg.UseGamescope {
		sb.WriteString(fmt.Sprintf("- **Gamescope**: %s (%dx%d@%dHz)\n", cfg.GamescopeOutput, cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeRefresh))
	}
	if cfg.UsePrimeRun {
		sb.WriteString("- **GPU Offload**: prime-run enabled\n")
	}
	if cfg.UsePCores {
		sb.WriteString(fmt.Sprintf("- **CPU Pinning**: P-cores mask `%s`\n", cfg.PCoresMask))
	}
	if len(cfg.DLLOverrides) > 0 {
		var ovList []string
		for k, v := range cfg.DLLOverrides {
			ovList = append(ovList, fmt.Sprintf("%s=%s", k, v))
		}
		sb.WriteString(fmt.Sprintf("- **DLL Overrides**: `%s`\n", strings.Join(ovList, ", ")))
	}
	if len(cfg.EnvVars) > 0 {
		var envKeys []string
		for k := range cfg.EnvVars {
			envKeys = append(envKeys, k)
		}
		sb.WriteString(fmt.Sprintf("- **Custom Environment Variables**: `%s`\n", strings.Join(envKeys, ", ")))
	}
	if cfg.PreLaunchHook != "" {
		sb.WriteString("- **Lifecycle Hook**: Custom pre_launch.sh script applied\n")
	}
	if !cfg.UseGamescope && !cfg.UsePrimeRun && !cfg.UsePCores && len(cfg.DLLOverrides) == 0 && len(cfg.EnvVars) == 0 && cfg.PreLaunchHook == "" {
		sb.WriteString("- **Tweaks**: None. Ran out of the box with default settings.\n")
	}

	// Session & Empirical Summary
	sb.WriteString("\n#### 📊 Session & Stability Summary\n")
	if opts.LastResult != nil {
		durationStr := opts.LastResult.Duration.Round(100000000).String()
		if opts.LastResult.CrashDetected {
			sb.WriteString(fmt.Sprintf("- **Last Session**: Crash detected after %s (Exit Code %d)\n", durationStr, opts.LastResult.ExitCode))
		} else {
			sb.WriteString(fmt.Sprintf("- **Last Session**: Clean exit after %s (Exit Code %d)\n", durationStr, opts.LastResult.ExitCode))
		}
	}
	if len(opts.History) > 0 {
		totalRuns := len(opts.History)
		cleanRuns := 0
		for _, h := range opts.History {
			if !h.CrashDetected && (h.ExitCode == 0 || h.AbortedByUser) {
				cleanRuns++
			}
		}
		sb.WriteString(fmt.Sprintf("- **Local History**: %d of %d recorded sessions ran smoothly\n", cleanRuns, totalRuns))
	}

	sb.WriteString(fmt.Sprintf("\n#### 🏆 Recommended Verdict: %s\n", strings.ToUpper(tier)))
	switch tier {
	case "Platinum":
		sb.WriteString("Runs flawlessly out of the box with zero tweaks or crashes required.\n")
	case "Gold":
		sb.WriteString("Runs completely smooth and stable once recommended runner and quirks are configured.\n")
	case "Silver":
		sb.WriteString("Playable, but encounters minor hiccups, intermittent freezes, or non-zero exits.\n")
	case "Bronze":
		sb.WriteString("Runs into frequent crashes, stuttering, or severe glitches.\n")
	case "Borked":
		sb.WriteString("Will not launch, crashes immediately, or blocked by anti-cheat/wine compatibility.\n")
	}

	sb.WriteString("\n---\n*Report generated with [run-proton-tui](https://github.com/broli/run-proton-tui)*\n")

	return sb.String()
}
