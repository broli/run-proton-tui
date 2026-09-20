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
	Width           int
	Height          int
}

func lbl(name string) string {
	return style.LabelStyle.Render(fmt.Sprintf("%-16s", name))
}

// RenderDashboard draws the responsive, high-readability overview dashboard.
func RenderDashboard(d DashboardData) string {
	// 1. Calculate responsive layout bounds
	termWidth := d.Width
	if termWidth <= 0 {
		termWidth = 110
	}
	contentWidth := termWidth - 4
	if contentWidth < 80 {
		contentWidth = 80
	} else if contentWidth > 140 {
		contentWidth = 140
	}

	colWidth := (contentWidth - 2) / 2
	rightColWidth := contentWidth - colWidth - 2

	var sb strings.Builder

	// 2. Full-Width Header Bar
	leftHeader := fmt.Sprintf("🎮 rpt v2.0.0 │ Game: %s", d.GameTitle)
	rightHeader := "[?] Help & Manual  [q] Quit"
	spaceCount := contentWidth - lipgloss.Width(leftHeader) - lipgloss.Width(rightHeader) - 2
	if spaceCount < 2 {
		spaceCount = 2
	}
	headerLine := leftHeader + strings.Repeat(" ", spaceCount) + rightHeader
	sb.WriteString(style.HeaderBar.Width(contentWidth).Render(headerLine))
	sb.WriteString("\n")

	// Status Message Banner if present
	if d.StatusMessage != "" {
		sb.WriteString(style.BadgeWarning.Render(" ℹ "+d.StatusMessage) + "\n\n")
	}

	// 3. Left Panel: Target & Compatibility
	var leftSb strings.Builder
	leftSb.WriteString(style.SectionTitle.Render("🎯 Target Executable & Compatibility") + "\n\n")

	// Executable & Classification Tag
	exeTag := style.BadgeHighlight.Render("3D Game Engine")
	if d.Classification.Type == runner.ExeType2DUtility {
		exeTag = style.BadgeWarning.Render("2D Utility / Launcher")
	}
	leftSb.WriteString(fmt.Sprintf("%s %s %s\n", lbl("Executable:"), style.KeyStyle.Render(d.Config.TargetExe), exeTag))

	// Rationale (wrapped cleanly to panel width)
	rationaleStyle := style.SubheaderStyle.Width(colWidth - 20)
	leftSb.WriteString(fmt.Sprintf("%s %s\n\n", lbl("Target Profile:"), rationaleStyle.Render(d.Classification.Rationale)))

	// Proton Version
	leftSb.WriteString(fmt.Sprintf("%s %s\n", lbl("Proton Runner:"), style.KeyStyle.Render(d.ProtonName)))

	// ProtonDB Rating
	if d.ProtonDB != nil {
		leftSb.WriteString(fmt.Sprintf("%s %s (%s, %d reports)\n\n",
			lbl("ProtonDB Tier:"),
			d.ProtonDB.GetTierBadge(),
			d.ProtonDB.Confidence,
			d.ProtonDB.Total))
	} else {
		leftSb.WriteString(fmt.Sprintf("%s %s\n\n",
			lbl("ProtonDB Tier:"),
			style.SubheaderStyle.Render("Unknown (Press [a] to fetch community ratings)")))
	}

	// Wine Prefix Status
	pfxInfo := "Active"
	if d.PrefixSize != "" {
		pfxInfo = fmt.Sprintf("Active (%s)", d.PrefixSize)
	}
	leftSb.WriteString(fmt.Sprintf("%s %s\n\n", lbl("Wine Prefix:"), fmt.Sprintf("./proton-prefix/ (%s)", pfxInfo)))

	// Steam Emulator State
	if d.Emulator != nil && d.Emulator.Detected {
		leftSb.WriteString(fmt.Sprintf("%s %s (AppID: %s)\n",
			lbl("Steam Emulator:"),
			style.ValueStyle.Render(d.Emulator.Type),
			style.KeyStyle.Render(d.Emulator.AppID)))
		leftSb.WriteString(fmt.Sprintf("%s %s\n",
			lbl("DLC Status:"),
			style.BadgeSuccess.Render(d.Emulator.DLCStatus)))
	} else {
		leftSb.WriteString(fmt.Sprintf("%s %s\n",
			lbl("Steam Emulator:"),
			style.SubheaderStyle.Render("No custom emulator detected (AppID: 0)")))
	}

	// 4. Right Panel: Hardware & Execution Stack
	var rightSb strings.Builder
	rightSb.WriteString(style.SectionTitle.Render("⚡ Hardware & Execution Pipeline") + "\n\n")

	renderRow := func(label, status string) {
		rightSb.WriteString(fmt.Sprintf("%s %s\n", lbl(label), status))
	}

	// Gamescope
	gsStatus := style.BadgeMuted.Render("Disabled (Native Window)")
	if d.Config.UseGamescope {
		gsStatus = style.BadgeSuccess.Render(fmt.Sprintf("ON (1080p @ %dHz -> %s)", d.Config.GamescopeRefresh, d.Config.GamescopeOutput))
	}
	renderRow("Gamescope:", gsStatus)

	// CPU P-Cores
	pcoreStatus := style.BadgeMuted.Render("All Cores (Default)")
	if d.Config.UsePCores {
		pcoreStatus = style.BadgeSuccess.Render(fmt.Sprintf("PINNED (Threads %s)", d.Config.PCoresMask))
	}
	renderRow("CPU Topology:", pcoreStatus)

	// GPU Offload
	gpuStatus := style.BadgeMuted.Render("Host iGPU")
	if d.Config.UsePrimeRun && d.HasPrimeRun {
		gpuStatus = style.BadgeSuccess.Render("prime-run (NVIDIA RTX 4060)")
	}
	renderRow("GPU Offload:", gpuStatus)

	// NT Sync
	syncStatus := style.BadgeHighlight.Render("FSYNC (futex)")
	if hardware.HasNTSync() {
		syncStatus = style.BadgeSuccess.Render("/dev/ntsync (Kernel Fast)")
	}
	renderRow("Sync Engine:", syncStatus)

	// Power Mode
	pwrStatus := style.BadgeMuted.Render("System Default")
	if d.Config.ManagePower {
		pwrStatus = style.BadgeSuccess.Render("Auto Performance (Reverts on exit)")
	}
	renderRow("Power Profile:", pwrStatus)

	// Xalia
	xaliaStatus := style.BadgeSuccess.Render("Disabled (Clean DXVK)")
	if d.Config.UseXalia {
		xaliaStatus = style.BadgeWarning.Render("Enabled (UI Automation)")
	}
	renderRow("Xalia Bridge:", xaliaStatus)

	// DLL Overrides
	overrideStr := style.SubheaderStyle.Render("None (Default)")
	if len(d.ActiveOverrides) > 0 {
		var oList []string
		for k, v := range d.ActiveOverrides {
			oList = append(oList, fmt.Sprintf("%s=%s", k, v))
		}
		overrideStr = style.ValueStyle.Render(strings.Join(oList, ", "))
	}
	renderRow("DLL Overrides:", overrideStr)

	// Logging
	logStatus := style.SubheaderStyle.Render("Disabled")
	if d.Config.EnableLogging {
		logStatus = style.BadgeSuccess.Render("Active (writing to .logs/)")
	}
	renderRow("Session Logs:", logStatus)

	// Render panels with matched height
	rawLeft := style.PanelStyle.Width(colWidth).Render(leftSb.String())
	rawRight := style.PanelStyle.Width(rightColWidth).Render(rightSb.String())

	hLeft := lipgloss.Height(rawLeft)
	hRight := lipgloss.Height(rawRight)
	targetH := hLeft
	if hRight > targetH {
		targetH = hRight
	}
	// Subtract 2 for top and bottom border
	panelInnerH := targetH - 2
	if panelInnerH < 1 {
		panelInnerH = 1
	}

	leftPanel := style.PanelStyle.Width(colWidth).Height(panelInnerH).Render(leftSb.String())
	rightPanel := style.PanelStyle.Width(rightColWidth).Height(panelInnerH).Render(rightSb.String())

	// Join both panels side-by-side with matched height
	panelsRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	sb.WriteString(panelsRow)
	sb.WriteString("\n")

	// 5. Command Preview Box (Full Width)
	cmdPreview := buildPreviewCommand(d)
	cmdHeader := style.SectionTitle.Render("🚀 Execution Pipeline Preview:") + "\n"
	cmdContent := lipgloss.NewStyle().Width(contentWidth - 4).Render(cmdPreview)
	sb.WriteString(style.CommandBoxStyle.Width(contentWidth).Render(cmdHeader + cmdContent))
	sb.WriteString("\n")

	// 6. Action Menu & Hotkeys Dock (Full Width, Balanced Rows)
	sb.WriteString(renderMenuDock(contentWidth))

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

func pill(key, label string) string {
	return style.KeyBadge.Render(key) + " " + style.KeyDesc.Render(label) + "  "
}

func renderMenuDock(width int) string {
	row1 := pill("[Enter]", "Launch Game") +
		pill("[2]", "Proton Runner") +
		pill("[3]", "Executable") +
		pill("[a]", "ProtonDB Tips")

	row2 := pill("[g]", "Gamescope") +
		pill("[p]", "P-Cores") +
		pill("[v]", "GPU Runner") +
		pill("[o]", "DLL Overrides")

	row3 := pill("[c]", "Reset Prefix") +
		pill("[h]", "Health Check") +
		pill("[l]", "Logs") +
		pill("[?]", "Help Manual") +
		pill("[q]", "Quit")

	dockContent := style.SectionTitle.Render("⌨️  Controls & Hotkeys:") + "\n" + row1 + "\n" + row2 + "\n" + row3

	return style.PanelStyle.Width(width).Render(dockContent)
}
