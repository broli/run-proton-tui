package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/runner"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/lipgloss"
)

// DashboardData provides all dynamic state required to render the main overview screen.
type DashboardData struct {
	GameTitle       string
	GameDir         string
	Config          *config.GameConfig
	Emulator        *integrations.EmulatorInfo
	ProtonDB        *integrations.ProtonDBReport
	Classification  runner.ClassificationResult
	HasNTSync       bool
	HasPrimeRun     bool
	HasGamescope    bool
	ActiveOverrides map[string]string
	PrefixSize      string
	ProtonName      string
	StatusMessage   string
}

// RenderDashboard draws the main overview dashboard.
func RenderDashboard(d DashboardData) string {
	var sb strings.Builder

	// 1. Title Bar
	title := fmt.Sprintf("🎮 rpt (Run Proton TUI) v2.0 — %s", d.GameTitle)
	sb.WriteString(style.TitleStyle.Render(title))
	sb.WriteString("\n")

	// Status Message Banner if present
	if d.StatusMessage != "" {
		sb.WriteString(style.BadgeWarning.Render(d.StatusMessage))
		sb.WriteString("\n\n")
	}

	// 2. Overview Grid (Left column: Game & Runner, Right column: Hardware & Emulators)
	var leftCol, rightCol strings.Builder

	// Game Target & Classification
	leftCol.WriteString(style.HeaderStyle.Render("Target Executable:\n"))
	exeBadge := style.BadgeHighlight.Render("3D Game")
	if d.Classification.Type == runner.ExeType2DUtility {
		exeBadge = style.BadgeWarning.Render("2D Utility / Launcher")
	}
	leftCol.WriteString(fmt.Sprintf("  %s %s\n", style.KeyStyle.Render(d.Config.TargetExe), exeBadge))
	leftCol.WriteString(fmt.Sprintf("  %s\n\n", style.SubheaderStyle.Render(d.Classification.Rationale)))

	// Proton Version
	leftCol.WriteString(style.HeaderStyle.Render("Compatibility Runner:\n"))
	leftCol.WriteString(fmt.Sprintf("  %s\n", style.KeyStyle.Render(d.ProtonName)))
	if d.ProtonDB != nil {
		leftCol.WriteString(fmt.Sprintf("  ProtonDB: %s (%s, %d reports)\n\n",
			style.BadgeSuccess.Render(d.ProtonDB.GetTierBadge()),
			d.ProtonDB.Confidence,
			d.ProtonDB.Total))
	} else {
		leftCol.WriteString(fmt.Sprintf("  %s\n\n", style.SubheaderStyle.Render("Press [a] to query ProtonDB summary")))
	}

	// Wine Prefix Status
	leftCol.WriteString(style.HeaderStyle.Render("Wine Prefix:\n"))
	pfxInfo := "Active"
	if d.PrefixSize != "" {
		pfxInfo = fmt.Sprintf("Active (%s)", d.PrefixSize)
	}
	leftCol.WriteString(fmt.Sprintf("  ./proton-prefix/ (%s)\n\n", pfxInfo))

	// Hardware & Environment (Right Column)
	rightCol.WriteString(style.HeaderStyle.Render("Execution & Hardware Stack:\n"))

	// Gamescope
	gsStatus := style.BadgeMuted.Render("Disabled")
	if d.Config.UseGamescope {
		gsStatus = style.BadgeSuccess.Render(fmt.Sprintf("1080p @ %dHz -> %s", d.Config.GamescopeRefresh, d.Config.GamescopeOutput))
	}
	rightCol.WriteString(fmt.Sprintf("  • Gamescope   : %s\n", gsStatus))

	// CPU P-Cores
	pcoreStatus := style.BadgeMuted.Render("All Cores")
	if d.Config.UsePCores {
		pcoreStatus = style.BadgeSuccess.Render(fmt.Sprintf("Pinned (Threads %s)", d.Config.PCoresMask))
	}
	rightCol.WriteString(fmt.Sprintf("  • CPU Cores   : %s\n", pcoreStatus))

	// GPU Runner
	gpuStatus := style.BadgeMuted.Render("Host iGPU")
	if d.Config.UsePrimeRun && d.HasPrimeRun {
		gpuStatus = style.BadgeSuccess.Render("prime-run (NVIDIA RTX 4060)")
	}
	rightCol.WriteString(fmt.Sprintf("  • GPU Runner  : %s\n", gpuStatus))

	// Fast Sync
	syncStatus := style.BadgeHighlight.Render("FSYNC (futex)")
	if hardware.HasNTSync() {
		syncStatus = style.BadgeSuccess.Render("/dev/ntsync (Kernel Fast)")
	}
	rightCol.WriteString(fmt.Sprintf("  • Sync Engine : %s\n", syncStatus))

	// Power Profile
	pwrStatus := style.BadgeMuted.Render("System Default")
	if d.Config.ManagePower {
		pwrStatus = style.BadgeSuccess.Render("Auto Performance (Reverts on exit)")
	}
	rightCol.WriteString(fmt.Sprintf("  • Power Mode  : %s\n\n", pwrStatus))

	// Steam Emulator Status
	rightCol.WriteString(style.HeaderStyle.Render("Steam Emulator & DLC State:\n"))
	if d.Emulator != nil && d.Emulator.Detected {
		rightCol.WriteString(fmt.Sprintf("  • Emulator : %s (AppID: %s)\n", style.BadgeHighlight.Render(d.Emulator.Type), d.Emulator.AppID))
		rightCol.WriteString(fmt.Sprintf("  • DLCs     : %s\n", style.BadgeSuccess.Render(d.Emulator.DLCStatus)))
	} else {
		rightCol.WriteString("  • No custom Steam emulator detected (AppID: 0)\n")
	}

	// Overrides summary
	if len(d.ActiveOverrides) > 0 {
		var oList []string
		for k, v := range d.ActiveOverrides {
			oList = append(oList, fmt.Sprintf("%s=%s", k, v))
		}
		rightCol.WriteString(fmt.Sprintf("  • Overrides: %s\n", strings.Join(oList, ", ")))
	}

	// Join Columns
	cols := lipgloss.JoinHorizontal(lipgloss.Top,
		lipgloss.NewStyle().Width(45).Render(leftCol.String()),
		lipgloss.NewStyle().Width(45).Render(rightCol.String()),
	)
	sb.WriteString(style.CardStyle.Render(cols))
	sb.WriteString("\n")

	// 3. Command Preview
	cmdPreview := buildPreviewCommand(d)
	sb.WriteString(style.CommandBoxStyle.Render("Launch Command: " + cmdPreview))
	sb.WriteString("\n\n")

	// 4. Action Menu & Hotkeys
	sb.WriteString(renderMenuHotkeys())

	return sb.String()
}

func buildPreviewCommand(d DashboardData) string {
	prefix := ""
	if d.Config.UsePrimeRun && d.HasPrimeRun {
		prefix = "prime-run "
	}
	if d.Config.UsePCores && d.Config.PCoresMask != "" {
		prefix = fmt.Sprintf("taskset -c %s %s", d.Config.PCoresMask, prefix)
	}

	inner := fmt.Sprintf("%s%s waitforexitandrun ./%s", prefix, d.ProtonName, d.Config.TargetExe)
	if d.Config.UseGamescope && d.HasGamescope {
		return fmt.Sprintf("gamescope --prefer-output %s -W %d -H %d -r %d -f -- %s",
			d.Config.GamescopeOutput, d.Config.GamescopeWidth, d.Config.GamescopeHeight, d.Config.GamescopeRefresh, inner)
	}
	return inner
}

func renderMenuHotkeys() string {
	keys := []string{
		style.KeyStyle.Render("[Enter/1] Launch Game"),
		style.KeyStyle.Render("[2] Select Proton"),
		style.KeyStyle.Render("[3] Select Executable"),
		style.KeyStyle.Render("[g] Toggle Gamescope"),
		style.KeyStyle.Render("[p] Toggle P-Cores"),
		style.KeyStyle.Render("[v] Toggle GPU Runner"),
		style.KeyStyle.Render("[o] DLL Overrides"),
		style.KeyStyle.Render("[a] ProtonDB Tips"),
		style.KeyStyle.Render("[c] Clean Prefix"),
		style.KeyStyle.Render("[h] Health Diagnostics"),
		style.KeyStyle.Render("[l] View Logs"),
		style.KeyStyle.Render("[?] In-Depth Help"),
		style.KeyStyle.Render("[q] Quit"),
	}

	return strings.Join(keys, "  ")
}
