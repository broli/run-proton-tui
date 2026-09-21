package runner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// DiagnosticInsight represents an automated deduction from the game session log.
type DiagnosticInsight struct {
	Category       string
	Observation    string
	Recommendation string
}

// AnalyzeSessionLog scans the captured log for known failure signatures.
func AnalyzeSessionLog(logPath string, cfg *config.GameConfig) []DiagnosticInsight {
	var insights []DiagnosticInsight

	f, err := os.Open(logPath)
	if err != nil {
		return insights
	}
	defer f.Close()

	hasGamescope := false
	hasCODA := false
	hasWineServerCrash := false
	hasDXVKInit := false
	hasVulkanError := false
	hasKeycodeClip := false
	hasFailedToOpen := false

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "[gamescope]") {
			hasGamescope = true
		}
		if strings.Contains(line, "CODA transfers") || strings.Contains(line, "binkw32") {
			hasCODA = true
		}
		if strings.Contains(line, "backtrace") || strings.Contains(line, "page fault") || strings.Contains(line, "SIGSEGV") {
			hasWineServerCrash = true
		}
		if strings.Contains(line, "info:   DXVK:") || strings.Contains(line, "info:   D3D9:") {
			hasDXVKInit = true
		}
		if strings.Contains(line, "VK_ERROR") || (strings.Contains(line, "vulkan") && strings.Contains(line, "failed")) {
			hasVulkanError = true
		}
		if strings.Contains(line, "xkbcomp") && strings.Contains(line, "clipping") {
			hasKeycodeClip = true
		}
		if strings.Contains(line, "failed to open") {
			hasFailedToOpen = true
		}
	}

	// 0. Executable not found in working directory
	if hasFailedToOpen {
		insights = append(insights, DiagnosticInsight{
			Category:       "Executable Not Found (Exit Code 53)",
			Observation:    fmt.Sprintf("Wine failed to open %q. The file does not exist in this directory.", cfg.TargetExe),
			Recommendation: "Verify your current working directory. You may have launched rpt in the wrong game directory, or misspelled the executable name.",
		})
	}

	// 1. Gamescope with Direct3D 9 titles
	if (hasGamescope || cfg.UseGamescope) && strings.Contains(strings.ToLower(cfg.TargetExe), "justcause") {
		insights = append(insights, DiagnosticInsight{
			Category:       "Gamescope & 32-bit Direct3D 9 Stall",
			Observation:    "Gamescope was active on Intel iGPU while game launched via prime-run. 32-bit DirectX 9 titles frequently deadlock when presenting swapchain frames to Gamescope's nested Xwayland server.",
			Recommendation: "Toggle Gamescope OFF ([g] in rpt or pass '--gamescope false'). Vintage D3D9 games run smoother natively under KWin Wayland/Xwayland.",
		})
	}

	// 2. Bink Video Codec / Splash Loading Bar Freeze
	if hasCODA {
		insights = append(insights, DiagnosticInsight{
			Category:       "Splash Screen & Bink Video / CODA Stall",
			Observation:    "Game displayed 2D splash window with loading bar, but froze while loading archives and allocating 64MB CODA buffer for intro video decompression before 3D rendering initialized.",
			Recommendation: "Toggle Gamescope OFF ('rpt --gamescope false --now'). Running 32-bit DirectX 9 directly on host KWin Xwayland allows the 2D splash window to cleanly hand off to the fullscreen Direct3D 9 engine. Also run 'rpt JCSetup.exe' in the Just Cause directory to lock 1080p resolution.",
		})
	}

	// 3. AppID mismatch for Just Cause 1 vs 3
	if (cfg.AppID == "225540" || cfg.AppID == "0" || cfg.AppID == "") && strings.Contains(strings.ToLower(cfg.TargetExe), "justcause") {
		insights = append(insights, DiagnosticInsight{
			Category:       "Steam AppID Mismatch",
			Observation:    fmt.Sprintf("Configured AppID is %q (Steam ID for Just Cause 3, 2015), but this is Just Cause 1 (2006).", cfg.AppID),
			Recommendation: "Set app_id = '6880' in .proton-config.toml so ProtonDB quirks and UMU fixes match the actual game.",
		})
	}

	// 4. Wine segmentation fault
	if hasWineServerCrash {
		insights = append(insights, DiagnosticInsight{
			Category:       "Wine Segmentation Fault",
			Observation:    "Crash dump or page fault detected in Wine process log.",
			Recommendation: "Check DLL overrides or test with another Proton runner (e.g. Proton 9.0 or GE-Proton9-23).",
		})
	}

	// 5. Vulkan driver error
	if hasVulkanError {
		insights = append(insights, DiagnosticInsight{
			Category:       "Vulkan Driver Error",
			Observation:    "Vulkan initialization failed in DXVK/Wine.",
			Recommendation: "Check GPU drivers or verify prime-run configuration for 32-bit Vulkan ICDs.",
		})
	}

	// 6. Generic Gamescope reminder if Gamescope was active without DXVK init
	if (hasGamescope || cfg.UseGamescope) && !hasDXVKInit && !hasCODA && len(insights) == 0 {
		insights = append(insights, DiagnosticInsight{
			Category:       "Gamescope Display Sandbox",
			Observation:    "Gamescope initialized display output, but game rendering pipeline never initialized.",
			Recommendation: "Test launching with Gamescope OFF: 'rpt --gamescope false --now'.",
		})
	}

	_ = hasKeycodeClip
	return insights
}

// ReadLogTail reads the last n lines of a file.
func ReadLogTail(logPath string, n int) []string {
	f, err := os.Open(logPath)
	if err != nil {
		return nil
	}
	defer f.Close()

	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > n {
			lines = lines[1:]
		}
	}
	return lines
}

// GenerateAgentPrompt formats a prompt that the user can copy and paste into any AI agent.
func GenerateAgentPrompt(result *SessionResult, cfg *config.GameConfig, insights []DiagnosticInsight, logTail []string) string {
	var sb strings.Builder

	sb.WriteString("### Game Launch & Crash Debug Context\n\n")
	sb.WriteString(fmt.Sprintf("- **Target Executable**: `%s`\n", cfg.TargetExe))
	sb.WriteString(fmt.Sprintf("- **Proton Runner**: `%s`\n", filepath.Base(cfg.ProtonPath)))
	sb.WriteString(fmt.Sprintf("- **Steam AppID**: `%s`\n", cfg.AppID))
	sb.WriteString(fmt.Sprintf("- **Gamescope**: `%v` (Output: %s, %dx%d@%dHz)\n",
		cfg.UseGamescope, cfg.GamescopeOutput, cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeRefresh))
	sb.WriteString(fmt.Sprintf("- **Prime-Run (NVIDIA)**: `%v`\n", cfg.UsePrimeRun))
	sb.WriteString(fmt.Sprintf("- **P-Cores Pinning**: `%v` (Mask: `%s`)\n", cfg.UsePCores, cfg.PCoresMask))
	sb.WriteString(fmt.Sprintf("- **Session Duration**: `%v` | **Exit Code**: `%d`\n", result.Duration.Round(100000000), result.ExitCode))
	if result.AbortedByUser {
		sb.WriteString("- **Termination**: Manually interrupted by user (Ctrl+C) because game appeared stuck / not loading.\n")
	} else if result.CrashDetected {
		sb.WriteString("- **Termination**: Game crashed or exited unexpectedly.\n")
	}

	if len(insights) > 0 {
		sb.WriteString("\n#### Automated Diagnostic Findings\n")
		for _, ins := range insights {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n  *Recommended Fix*: %s\n", ins.Category, ins.Observation, ins.Recommendation))
		}
	}

	if len(logTail) > 0 {
		sb.WriteString("\n#### Session Log (Last lines before termination)\n```text\n")
		for _, l := range logTail {
			sb.WriteString(l + "\n")
		}
		sb.WriteString("```\n")
	}

	sb.WriteString("\n#### Question for Agent\n")
	sb.WriteString("Why did this game hang or fail to load with these settings, and what exact configuration or command should I use in Linux / Proton to get it running smoothly?\n")

	return sb.String()
}

// FormatDiagnosticReport renders a visually structured, user-friendly diagnostic card without awkward wrapping.
func FormatDiagnosticReport(result *SessionResult, cfg *config.GameConfig, insights []DiagnosticInsight) string {
	boxWidth := 94

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorDanger)

	quoteStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(style.ColorHighlight)

	badgeStyle := lipgloss.NewStyle().
		Bold(true).
		Background(style.ColorWarning).
		Foreground(lipgloss.Color("#1a1b26")).
		Padding(0, 1)

	headerStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary)

	textStyle := lipgloss.NewStyle().
		Foreground(style.ColorText)

	mutedStyle := lipgloss.NewStyle().
		Foreground(style.ColorMuted)

	fixStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorSuccess)

	var lines []string

	if result.AbortedByUser {
		lines = append(lines, titleStyle.Render("⚠️  SESSION TERMINATED BY USER (Ctrl+C) / GAME HANG DETECTED"))
		lines = append(lines, quoteStyle.Render("“We believe the game crashed, did not work, or was terminated by you because of an issue.”"))
	} else {
		lines = append(lines, titleStyle.Render(fmt.Sprintf("💥 GAME CRASH DETECTED (Exit Code: %d after %v)", result.ExitCode, result.Duration.Round(100000000))))
		lines = append(lines, quoteStyle.Render("“We believe the game crashed, failed to initialize, or encountered a fatal error.”"))
	}
	lines = append(lines, "")

	// Session details
	lines = append(lines, headerStyle.Render("Session Information:"))
	lines = append(lines, textStyle.Render(fmt.Sprintf("  • Target: %s  |  Runner: %s", cfg.TargetExe, filepath.Base(cfg.ProtonPath))))
	lines = append(lines, textStyle.Render(fmt.Sprintf("  • Gamescope: %v  |  Prime-Run: %v  |  AppID: %s", cfg.UseGamescope, cfg.UsePrimeRun, cfg.AppID)))
	if result.LogFile != "" {
		lines = append(lines, textStyle.Render(fmt.Sprintf("  • Session Log: %s", result.LogFile)))
	}
	lines = append(lines, "")

	// Diagnostic Insights
	if len(insights) > 0 {
		lines = append(lines, headerStyle.Render("Automated Root-Cause Insights:"))
		for _, ins := range insights {
			lines = append(lines, fmt.Sprintf("  %s %s", badgeStyle.Render(ins.Category), textStyle.Render(ins.Observation)))
			lines = append(lines, fmt.Sprintf("    %s %s", fixStyle.Render("Fix:"), textStyle.Render(ins.Recommendation)))
		}
		lines = append(lines, "")
	} else {
		lines = append(lines, headerStyle.Render("Analysis:"))
		lines = append(lines, textStyle.Render("  No standard failure signatures matched in the log."))
		lines = append(lines, "")
	}

	// Dynamic Actions based on context
	hasMissingExe := false
	for _, ins := range insights {
		if strings.Contains(ins.Category, "Executable Not Found") {
			hasMissingExe = true
			break
		}
	}

	isJustCause := strings.Contains(strings.ToLower(cfg.TargetExe), "justcause") || strings.Contains(strings.ToLower(cfg.TargetExe), "jcsetup")

	if hasMissingExe {
		lines = append(lines, headerStyle.Render("Recommended Action:"))
		lines = append(lines, textStyle.Render(fmt.Sprintf("  1. Verify working directory: %q does not exist in current folder.", cfg.TargetExe)))
		lines = append(lines, mutedStyle.Render("     Make sure you have navigated to the correct game directory before running rpt."))
		lines = append(lines, "")
	} else if isJustCause {
		lines = append(lines, headerStyle.Render("Quick Actions to Fix Just Cause:"))
		lines = append(lines, textStyle.Render("  1. Run graphics setup first (inside Just Cause folder):"))
		lines = append(lines, fixStyle.Render("     rpt JCSetup.exe"))
		lines = append(lines, mutedStyle.Render("     (Configure display resolution to 1920x1080 32-bit before running JustCause.exe)"))
		lines = append(lines, textStyle.Render("  2. Launch without Gamescope (DirectX 9 compatibility mode):"))
		lines = append(lines, fixStyle.Render("     rpt --gamescope false --now"))
		lines = append(lines, mutedStyle.Render("     (Bypasses Gamescope Xwayland presentation stall on hybrid Intel+NVIDIA)"))
		lines = append(lines, textStyle.Render("  3. Set correct Steam AppID in .proton-config.toml:"))
		lines = append(lines, fixStyle.Render("     app_id = '6880'"))
		lines = append(lines, mutedStyle.Render("     (Matches Just Cause 1 instead of Just Cause 3 for ProtonDB and UMU)"))
		lines = append(lines, "")
	} else {
		lines = append(lines, headerStyle.Render("Recommended Actions:"))
		lines = append(lines, textStyle.Render("  • Run health check: 'rpt --diagnostics' to inspect prefix permissions."))
		lines = append(lines, textStyle.Render("  • Test launching without Gamescope: 'rpt --gamescope false --now'."))
		lines = append(lines, textStyle.Render("  • Try a different Proton runner (e.g. Proton 9.0 or GE-Proton9-23)."))
		lines = append(lines, "")
	}

	// Copy-paste prompt guide
	lines = append(lines, headerStyle.Render("📋 Need Deeper Help? Agent-Ready Prompt Generated:"))
	lines = append(lines, mutedStyle.Render("  Inspect the agent prompt block below to copy-paste into an agent."))

	content := strings.Join(lines, "\n")

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWarning).
		Padding(1, 2).
		Width(boxWidth)

	return cardStyle.Render(content)
}
