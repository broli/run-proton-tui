package runner

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// DiagnosticInsight represents an automated deduction from the game session log.
type DiagnosticInsight struct {
	Category       string
	Observation    string
	Recommendation string
}

// ExitCodeExplanation provides plain-English translation and context for an exit code.
type ExitCodeExplanation struct {
	Title       string
	Description string
}

// ExplainExitCode translates process exit codes into descriptive diagnoses.
func ExplainExitCode(code int) ExitCodeExplanation {
	switch code {
	case 0:
		return ExitCodeExplanation{
			Title:       "Clean / Normal Exit (Code 0)",
			Description: "The process closed without an immediate shell error code. If the game closed prematurely, this often indicates an anti-cheat heartbeat timeout or display server window closure.",
		}
	case 1:
		return ExitCodeExplanation{
			Title:       "General Application Failure (Exit Code 1)",
			Description: "The executable or runner wrapper encountered an unhandled general runtime error.",
		}
	case 53:
		return ExitCodeExplanation{
			Title:       "Executable File Not Found (Exit Code 53)",
			Description: "Wine was unable to locate or open the specified Windows executable in the working directory.",
		}
	case 126:
		return ExitCodeExplanation{
			Title:       "Permission Denied (Exit Code 126)",
			Description: "The executable is missing execute permissions (+x) or the filesystem is mounted with noexec.",
		}
	case 127:
		return ExitCodeExplanation{
			Title:       "Binary / Dynamic Library Not Found (Exit Code 127)",
			Description: "A required command, interpreter, or runtime library could not be resolved in system PATH.",
		}
	case 130:
		return ExitCodeExplanation{
			Title:       "Session Terminated by User (SIGINT / Exit Code 130)",
			Description: "The session was manually interrupted by the user (Ctrl+C).",
		}
	case 134:
		return ExitCodeExplanation{
			Title:       "Process Aborted (SIGABRT / Exit Code 134)",
			Description: "The application or graphics driver aborted abnormally (e.g. glibc assertion failure or KWin abort).",
		}
	case 137:
		return ExitCodeExplanation{
			Title:       "Terminated by Out-Of-Memory Killer (SIGKILL / Exit Code 137)",
			Description: "The Linux kernel Out-Of-Memory (OOM) killer terminated the process due to system RAM/VRAM exhaustion.",
		}
	case 139:
		return ExitCodeExplanation{
			Title:       "Segmentation Fault (SIGSEGV / Exit Code 139)",
			Description: "The game or Wine driver attempted to read or write to an invalid memory address.",
		}
	case 222:
		return ExitCodeExplanation{
			Title:       "Application Crash Handler Trapped Fatal Exception (Exit Code 222)",
			Description: "An internal crash handler or runtime watchdog caught an unhandled fatal exception in the game or anti-cheat service and generated an error dump before exiting.",
		}
	default:
		if code > 128 && code <= 165 {
			sig := code - 128
			return ExitCodeExplanation{
				Title:       fmt.Sprintf("Terminated by Signal %d (Exit Code %d)", sig, code),
				Description: fmt.Sprintf("The application was terminated by Linux kernel signal %d.", sig),
			}
		}
		return ExitCodeExplanation{
			Title:       fmt.Sprintf("Non-Zero Return (Exit Code %d)", code),
			Description: fmt.Sprintf("The process finished with return code %d.", code),
		}
	}
}

// AnalyzeSessionLog scans the captured log for known failure signatures using built-in checks and modular dictionaries.
func AnalyzeSessionLog(logPath string, cfg *config.GameConfig, exitCodes ...int) []DiagnosticInsight {
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
	hasFailedToOpen := false

	gameDir := filepath.Dir(logPath)
	if filepath.Base(gameDir) == ".logs" {
		gameDir = filepath.Dir(gameDir)
	}
	signatures := LoadSignatures(gameDir)
	matchedSigCategories := make(map[string]bool)

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
		if strings.Contains(line, "info:   DXVK:") || strings.Contains(line, "info:   D3D9:") ||
			strings.Contains(line, "[Gamescope WSI] Made gamescope surface") ||
			strings.Contains(line, "[Gamescope WSI] Created swapchain") {
			hasDXVKInit = true
		}
		if strings.Contains(line, "VK_ERROR") || (strings.Contains(line, "vulkan") && strings.Contains(line, "failed")) {
			hasVulkanError = true
		}
		if strings.Contains(line, "failed to open") {
			hasFailedToOpen = true
		}

		// Proactive scan against modular signatures dictionary
		for _, sig := range signatures {
			if !matchedSigCategories[sig.Category] && MatchSignature(line, sig) {
				matchedSigCategories[sig.Category] = true
				insights = append(insights, DiagnosticInsight{
					Category:       sig.Category,
					Observation:    sig.Observation,
					Recommendation: sig.Recommendation,
				})
			}
		}
	}
	_ = scanner.Err()

	// Universal Built-in Diagnostics (benefit every game)
	// 0. Executable not found in working directory
	if hasFailedToOpen {
		insights = append(insights, DiagnosticInsight{
			Category:       "Executable Not Found (Exit Code 53)",
			Observation:    fmt.Sprintf("Wine failed to open %q. The file does not exist in this directory.", cfg.TargetExe),
			Recommendation: "Verify your current working directory. You may have launched rpt in the wrong game directory, or misspelled the executable name.",
		})
	}

	// 1. Gamescope with Direct3D 9 / vintage engines
	if (hasGamescope || cfg.UseGamescope) && strings.Contains(strings.ToLower(cfg.TargetExe), "justcause") {
		insights = append(insights, DiagnosticInsight{
			Category:       "Gamescope & 32-bit Direct3D 9 Stall",
			Observation:    "Gamescope was active on Intel iGPU while game launched via prime-run. 32-bit DirectX 9 titles frequently deadlock when presenting swapchain frames to Gamescope's nested Xwayland server.",
			Recommendation: "Consider testing with Gamescope disabled ('rpt --gamescope false --now'). Vintage D3D9 games may present cleaner frames natively under KWin Wayland/Xwayland.",
		})
	}

	// 2. Bink Video Codec / Splash Loading Bar Freeze (fallback if not already caught by signatures)
	if hasCODA && !matchedSigCategories["Bink Video / Intro Splash Stall"] {
		insights = append(insights, DiagnosticInsight{
			Category:       "Splash Screen & Bink Video / CODA Stall",
			Observation:    "Game displayed 2D splash window with loading bar, but froze while loading archives and allocating 64MB CODA buffer for intro video decompression before 3D rendering initialized.",
			Recommendation: "Testing with Gamescope disabled ('rpt --gamescope false --now') may allow the 2D splash window to cleanly hand off to the fullscreen Direct3D 9 engine.",
		})
	}

	// 3. AppID mismatch detection
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
			Category:       "Display Initialization",
			Observation:    "Gamescope initialized display output, but game rendering pipeline never initialized.",
			Recommendation: "Verify game prerequisites and runner compatibility, or test launching natively without Gamescope ('rpt --gamescope false --now') to check for background error dialogs.",
		})
	}

	// 7. Process Exit Code Insights
	if len(exitCodes) > 0 {
		code := exitCodes[0]
		if code == 222 && !matchedSigCategories["Crash Handler / Fatal Exception Trap"] {
			insights = append(insights, DiagnosticInsight{
				Category:       "Crash Handler Exception (Exit Code 222)",
				Observation:    "An internal crash handler caught an unhandled fatal exception during game startup and wrote a crash dump.",
				Recommendation: "Inspect runtime logs and crash dumps for stack traces, or ask an AI assistant to analyze the log.",
			})
		} else if code == 139 && !hasWineServerCrash {
			insights = append(insights, DiagnosticInsight{
				Category:       "Memory Segmentation Fault (Exit Code 139)",
				Observation:    "The process was terminated by SIGSEGV (invalid memory address dereference).",
				Recommendation: "Check DLL overrides, verify graphics drivers, or test with an alternate Proton runner.",
			})
		} else if code == 134 {
			insights = append(insights, DiagnosticInsight{
				Category:       "Process Abort (Exit Code 134 / SIGABRT)",
				Observation:    "The process was aborted abnormally by a runtime assertion failure or signal.",
				Recommendation: "Check session logs for glibc or graphics library abort messages, and verify compatibility settings.",
			})
		}
	}

	return insights
}

// ReadLogTail reads the last n lines of a file safely.
func ReadLogTail(logPath string, n int) []string {
	return diagnostics.ReadLogTail(logPath, n)
}

// GenerateAIHelperPackage formats a complete, self-contained diagnostic prompt for AI assistants (like ChatGPT, Claude, etc.).
func GenerateAIHelperPackage(result *SessionResult, cfg *config.GameConfig, insights []DiagnosticInsight, logTail []string, absLogPath string) string {
	var sb strings.Builder
	exp := ExplainExitCode(result.ExitCode)

	sb.WriteString("### 🎮 Linux Game Launch Diagnostic Helper\n\n")
	sb.WriteString("I am trying to run a Windows game on Linux using `rpt` (Run Proton TUI), but the game stopped unexpectedly or took too long to load.\n")
	sb.WriteString("Please review the system environment, game settings, and error logs below.\n\n")

	sb.WriteString("#### Game & System Overview\n")
	sb.WriteString(fmt.Sprintf("- **Target Executable**: `%s`\n", cfg.TargetExe))
	sb.WriteString(fmt.Sprintf("- **Proton Runner**: `%s`\n", filepath.Base(cfg.ProtonPath)))
	sb.WriteString(fmt.Sprintf("- **Steam AppID**: `%s`\n", cfg.AppID))
	sb.WriteString(fmt.Sprintf("- **Gamescope**: `%v` (Output: %s, %dx%d@%dHz)\n",
		cfg.UseGamescope, cfg.GamescopeOutput, cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeRefresh))
	sb.WriteString(fmt.Sprintf("- **Prime-Run (NVIDIA)**: `%v`\n", cfg.UsePrimeRun))
	sb.WriteString(fmt.Sprintf("- **P-Cores Pinning**: `%v` (Mask: `%s`)\n", cfg.UsePCores, cfg.PCoresMask))
	sb.WriteString(fmt.Sprintf("- **Session Duration**: `%v` | **Exit Code**: `%d` (%s)\n", result.Duration.Round(100*time.Millisecond), result.ExitCode, exp.Title))
	sb.WriteString(fmt.Sprintf("- **Exit Interpretation**: %s\n", exp.Description))
	if result.AbortedByUser {
		sb.WriteString("- **Termination**: Manually interrupted by user (Ctrl+C) because game appeared stuck / not loading.\n")
	} else if result.CrashDetected {
		sb.WriteString("- **Termination**: Game crashed, exited early, or failed to initialize.\n")
	}
	if absLogPath != "" {
		sb.WriteString(fmt.Sprintf("- **Full Log Path**: `%s`\n", absLogPath))
	}

	if len(insights) > 0 {
		sb.WriteString("\n#### Automated Diagnostic Findings\n")
		for _, ins := range insights {
			sb.WriteString(fmt.Sprintf("- **%s**: %s\n  *Proposed Solution*: %s\n", ins.Category, ins.Observation, ins.Recommendation))
		}
	}

	if len(logTail) > 0 {
		sb.WriteString("\n#### Session Log (Last lines before termination)\n```text\n")
		for _, l := range logTail {
			sb.WriteString(l + "\n")
		}
		sb.WriteString("```\n")
	}

	sb.WriteString("\n#### Official Reference Documentation for rpt\n")
	sb.WriteString("- Configuration Reference (`.proton-config.toml`): https://github.com/broli/run-proton-tui/wiki/Configuration-Reference\n")
	sb.WriteString("- Lifecycle Hooks & Workarounds (`hooks/pre_launch.sh`): https://github.com/broli/run-proton-tui/wiki/Lifecycle-Hooks-and-Preservation\n")
	sb.WriteString("- Real-World Case Study (Arknights: Endfield): https://github.com/broli/run-proton-tui/wiki/Example-Config-Arknights-Endfield\n")

	sb.WriteString("\n#### Instructions for the AI Assistant\n")
	sb.WriteString("1. First, check if the issue can be resolved with standard settings in `.proton-config.toml` (e.g. changing Proton runner, toggling Gamescope, adding environment variables, or configuring DLL overrides).\n")
	sb.WriteString("2. If the game requires custom fixes (such as binary patching, checking or patching Wine `ntoskrnl.exe`, memory mapping `/dev/shm`, or anti-cheat workarounds), **DO NOT tell the user to modify the rpt binary**. Instead, write a clean, idempotent bash script to be saved in `hooks/pre_launch.sh` (or `hooks/post_exit.sh`).\n")
	sb.WriteString("3. **DO NOT add Wine prefix flushing or wineserver kill/lock cleanup to hooks**: `rpt` already runs `wineserver -k`, `wineserver -w`, per-prefix orphan cleanup, and non-blocking stale socket lock sweeps (`/tmp/.wine-<UID>`) automatically BEFORE launch, inside the process wrapper, and AFTER exit. Hooks should NEVER call `wineserver -k` or delete Wine locks.\n")
	sb.WriteString("4. Provide step-by-step instructions so the user can just copy-paste your solution and run `rpt`!\n")
	sb.WriteString("5. If your solution works, remind the user that they can submit it to ProtonDB or as an `rpt` community hook to help other players!\n")

	sb.WriteString("\n#### Question for Agent\n")
	sb.WriteString("Why did this game hang or fail to load with these settings, and what exact configuration or command should I use in Linux / Proton to get it running smoothly?\n")

	return sb.String()
}

// GenerateAgentPrompt is a backwards-compatible wrapper calling GenerateAIHelperPackage.
func GenerateAgentPrompt(result *SessionResult, cfg *config.GameConfig, insights []DiagnosticInsight, logTail []string) string {
	return GenerateAIHelperPackage(result, cfg, insights, logTail, result.LogFile)
}

// FormatDiagnosticReport renders a visually structured, user-friendly diagnostic card with absolute file paths.
func FormatDiagnosticReport(result *SessionResult, cfg *config.GameConfig, insights []DiagnosticInsight, absHelpFilePath string) string {
	boxWidth := 94

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorDanger)

	quoteStyle := lipgloss.NewStyle().
		Italic(true).
		Foreground(style.ColorHighlight)

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
	exp := ExplainExitCode(result.ExitCode)

	if result.AbortedByUser {
		lines = append(lines, titleStyle.Render("⚠️  SESSION TERMINATED BY USER (Ctrl+C)"))
		lines = append(lines, quoteStyle.Render(fmt.Sprintf("Status: %s", exp.Description)))
	} else {
		lines = append(lines, titleStyle.Render(fmt.Sprintf("💥 PROCESS TERMINATED UNEXPECTEDLY (Exit Code: %d after %v)", result.ExitCode, result.Duration.Round(100*time.Millisecond))))
		lines = append(lines, quoteStyle.Render(fmt.Sprintf("Status: %s", exp.Description)))
	}
	lines = append(lines, "")

	// Session details
	lines = append(lines, headerStyle.Render("Session Information:"))
	lines = append(lines, textStyle.Render(fmt.Sprintf("  • Target: %s  |  Runner: %s", cfg.TargetExe, filepath.Base(cfg.ProtonPath))))
	lines = append(lines, textStyle.Render(fmt.Sprintf("  • Gamescope: %v  |  Prime-Run: %v  |  AppID: %s", cfg.UseGamescope, cfg.UsePrimeRun, cfg.AppID)))
	lines = append(lines, textStyle.Render(fmt.Sprintf("  • Exit Status: %s", exp.Title)))
	if result.LogFile != "" {
		lines = append(lines, textStyle.Render(fmt.Sprintf("  • Session Log: %s", result.LogFile)))
	}
	lines = append(lines, "")

	// Proposed Solution
	lines = append(lines, headerStyle.Render("Proposed Solution:"))
	if result.LogFile != "" {
		lines = append(lines, textStyle.Render(fmt.Sprintf("  1. Analyze session logs directly (%s) for the fatal error or assertion.", result.LogFile)))
	} else {
		lines = append(lines, textStyle.Render("  1. Analyze runtime logs in .logs/ for the fatal error or assertion."))
	}
	if absHelpFilePath != "" {
		lines = append(lines, textStyle.Render(fmt.Sprintf("  2. Or ask an AI assistant to analyze the logs using the pre-formatted report in %s.", absHelpFilePath)))
	} else {
		lines = append(lines, textStyle.Render("  2. Or ask an AI assistant to analyze the logs."))
	}
	lines = append(lines, "")

	// Troubleshooting Suggestions
	lines = append(lines, headerStyle.Render("Troubleshooting Suggestions:"))
	lines = append(lines, textStyle.Render("  • Check Logs: Review runtime logs in .logs/ for driver faults, missing DLLs, or memory errors."))

	searchQuery := cfg.PresetName
	if searchQuery == "" {
		searchQuery = strings.TrimSuffix(cfg.TargetExe, ".exe")
		if searchQuery == "" {
			searchQuery = filepath.Base(filepath.Dir(cfg.TargetExe))
		}
	}
	lines = append(lines, textStyle.Render(fmt.Sprintf("  • Search Google: Look up '%s proton crash' or '%s exit code %d' for known community fixes.", searchQuery, searchQuery, result.ExitCode)))
	lines = append(lines, textStyle.Render("  • Search Reddit: Ask or search in r/linux_gaming and r/SteamDeck with your game title and log snippet."))

	if cfg.HasValidAppID() {
		lines = append(lines, textStyle.Render(fmt.Sprintf("  • Check ProtonDB: https://www.protondb.com/app/%s for community ratings, tiers, and launch options.", cfg.AppID)))
	} else {
		lines = append(lines, textStyle.Render("  • Check ProtonDB: https://www.protondb.com to check if this title requires specific runner tweaks."))
	}
	lines = append(lines, textStyle.Render("  • Clean Zero Baseline: Run 'rpt --clean-zero --now' to test with pure upstream defaults and zero extra flags."))
	lines = append(lines, textStyle.Render("  • Run Health Check: 'rpt --diagnostics' to verify prefix permissions and file integrity."))
	lines = append(lines, textStyle.Render("  • Test Alternate Runner: Switch between GE-Proton, Proton Experimental, or Proton 9.0."))
	lines = append(lines, "")

	// AI Assistant Helper file instructions with absolute path
	if absHelpFilePath != "" {
		lines = append(lines, headerStyle.Render("📋 Ask AI to Analyze Logs:"))
		lines = append(lines, fixStyle.Render(fmt.Sprintf("  📄 Diagnostic file created at: %s", absHelpFilePath)))
		lines = append(lines, mutedStyle.Render("  1. Copy the contents of the file above."))
		lines = append(lines, mutedStyle.Render("  2. Paste it into your favorite AI assistant (like ChatGPT, Claude, etc.)."))
		lines = append(lines, quoteStyle.Render("  The AI assistant will analyze the logs and environment to propose targeted fixes."))
	}

	content := strings.Join(lines, "\n")

	cardStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWarning).
		Padding(1, 2).
		Width(boxWidth)

	return cardStyle.Render(content)
}
