package views

import (
	"fmt"
	"os"
	"path/filepath"
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
	OpaqueBackdrop  bool
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

	panelStyle := style.GetPanelStyle(d.OpaqueBackdrop)
	cmdBoxStyle := style.GetCommandBoxStyle(d.OpaqueBackdrop)
	headerBarStyle := style.GetHeaderBarStyle(d.OpaqueBackdrop)

	// 2. Full-Width Header Bar with Backdrop Toggle
	backdropText := "[b] Backdrop: Solid"
	if !d.OpaqueBackdrop {
		backdropText = "[b] Backdrop: Transparent"
	}
	leftHeader := fmt.Sprintf("🎮 rpt v0.5.0-alpha │ Game: %s", d.GameTitle)
	rightHeader := fmt.Sprintf("%s  •  [?] Help  •  [q] Quit", style.KeyStyle.Render(backdropText))
	spaceCount := contentWidth - lipgloss.Width(leftHeader) - lipgloss.Width(rightHeader) - 2
	if spaceCount < 2 {
		spaceCount = 2
	}
	headerLine := leftHeader + strings.Repeat(" ", spaceCount) + rightHeader
	sb.WriteString(headerBarStyle.Width(contentWidth).Render(headerLine))
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

	profileDesc := style.BadgeHighlight.Render("3D Game Engine (Dedicated GPU + Gamescope)")
	if d.Classification.Type == runner.ExeType2DUtility {
		profileDesc = style.BadgeWarning.Render("2D Utility / Launcher (Host iGPU Mode)")
	}
	if _, ok := d.Config.Profiles[d.Config.TargetExe]; ok {
		profileDesc = style.BadgeSuccess.Render("Custom Profile (.proton-config.toml)")
	}
	leftSb.WriteString(fmt.Sprintf("%s %s\n\n", lbl("Target Profile:"), profileDesc))

	// Proton Version
	leftSb.WriteString(fmt.Sprintf("%s %s\n", lbl("Proton Runner:"), style.KeyStyle.Render(d.ProtonName)))

	// ProtonDB Rating
	if d.ProtonDB != nil {
		leftSb.WriteString(fmt.Sprintf("%s %s (%s, %d reports) %s\n\n",
			lbl("ProtonDB Tier:"),
			d.ProtonDB.GetTierBadge(),
			d.ProtonDB.Confidence,
			d.ProtonDB.Total,
			style.KeyStyle.Render("([a] Info)")))
	} else {
		leftSb.WriteString(fmt.Sprintf("%s %s %s\n\n",
			lbl("ProtonDB Tier:"),
			style.SubheaderStyle.Render("Unknown"),
			style.KeyStyle.Render("([a] Fetch & View)")))
	}

	// Wine Prefix Status
	pfxDriveC := filepath.Join(d.GameDir, "proton-prefix", "pfx", "drive_c")
	pfxInfo := "Not Initialized (Created on launch)"
	if _, err := os.Stat(pfxDriveC); err == nil {
		if d.PrefixSize != "" {
			pfxInfo = fmt.Sprintf("Active (%s)", d.PrefixSize)
		} else {
			pfxInfo = "Active"
		}
	}
	leftSb.WriteString(fmt.Sprintf("%s %s\n\n", lbl("Wine Prefix:"), fmt.Sprintf("./proton-prefix/ (%s)", pfxInfo)))

	// Steam Emulator State
	if d.Emulator != nil && d.Emulator.Detected {
		leftSb.WriteString(fmt.Sprintf("%s %s (AppID: %s)\n",
			lbl("Steam Emulator:"),
			style.ValueStyle.Render(d.Emulator.Type),
			style.KeyStyle.Render(d.Emulator.AppID)))
		leftSb.WriteString(fmt.Sprintf("%s %s\n\n",
			lbl("DLC Status:"),
			style.BadgeSuccess.Render(d.Emulator.DLCStatus)))
	} else {
		leftSb.WriteString(fmt.Sprintf("%s %s\n\n",
			lbl("Steam Emulator:"),
			style.SubheaderStyle.Render("No custom emulator detected (AppID: 0)")))
	}

	// Quirks Preset Status (UMU & Custom Quirks)
	presetStatus := style.SubheaderStyle.Render("Standard Defaults ([d] Detect)")
	if d.Config.PresetName != "" {
		presetStatus = fmt.Sprintf("%s %s", style.BadgeSuccess.Render("⚡ "+d.Config.PresetName), style.KeyStyle.Render("([d] View)"))
	} else if d.Config.UmuID != "" && d.Config.UmuID != "umu-default" {
		presetStatus = fmt.Sprintf("%s %s", style.BadgeHighlight.Render("⚡ UMU: "+d.Config.UmuID), style.KeyStyle.Render("([d] View)"))
	}
	leftSb.WriteString(fmt.Sprintf("%s %s\n", lbl("Quirks Preset:"), presetStatus))

	// 4. Right Panel: Hardware & Execution Stack
	var rightSb strings.Builder
	rightSb.WriteString(style.SectionTitle.Render("⚡ Hardware & Execution Pipeline") + "\n\n")

	renderRow := func(label, status string) {
		rightSb.WriteString(fmt.Sprintf("%s %s\n", lbl(label), status))
	}

	// Gamescope
	gsOutput := d.Config.GamescopeOutput
	if gsOutput == "" || strings.EqualFold(gsOutput, "auto") {
		gsOutput = "Auto (External)"
	}
	gsStatus := style.BadgeMuted.Render("Disabled (Native Window)")
	if d.Config.UseGamescope {
		gsStatus = fmt.Sprintf("%s %s", style.BadgeSuccess.Render(fmt.Sprintf("ON (1080p@%dHz -> %s)", d.Config.GamescopeRefresh, gsOutput)), style.KeyStyle.Render("([m] Cycle)"))
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

	// DLL Overrides (Compact Yes / No indicator with summary)
	overrideStr := style.BadgeMuted.Render("No (None Active)")
	if len(d.ActiveOverrides) > 0 {
		var oList []string
		for k := range d.ActiveOverrides {
			oList = append(oList, k)
		}
		summary := strings.Join(oList, ", ")
		if len(summary) > 22 {
			summary = summary[:19] + "..."
		}
		overrideStr = style.BadgeSuccess.Render(fmt.Sprintf("Yes (%d active: %s)", len(d.ActiveOverrides), summary))
	}
	renderRow("DLL Overrides:", overrideStr)


	// Logging
	logStatus := fmt.Sprintf("%s %s", style.SubheaderStyle.Render("Disabled"), style.KeyStyle.Render("([L] Toggle)"))
	if d.Config.EnableLogging {
		logStatus = fmt.Sprintf("%s %s", style.BadgeSuccess.Render("Active (.logs/)"), style.KeyStyle.Render("([L] Toggle)"))
	}
	renderRow("Session Logs:", logStatus)

	// Render panels with matched height
	rawLeft := panelStyle.Width(colWidth).Render(leftSb.String())
	rawRight := panelStyle.Width(rightColWidth).Render(rightSb.String())

	hLeft := lipgloss.Height(rawLeft)
	hRight := lipgloss.Height(rawRight)
	targetH := hLeft
	if hRight > targetH {
		targetH = hRight
	}
	panelInnerH := targetH - 2
	if panelInnerH < 1 {
		panelInnerH = 1
	}

	leftPanel := panelStyle.Width(colWidth).Height(panelInnerH).Render(leftSb.String())
	rightPanel := panelStyle.Width(rightColWidth).Height(panelInnerH).Render(rightSb.String())

	// Join both panels side-by-side
	panelsRow := lipgloss.JoinHorizontal(lipgloss.Top, leftPanel, "  ", rightPanel)
	sb.WriteString(panelsRow)
	sb.WriteString("\n")

	// 5. Command Preview Box (Full Width)
	cmdPreview := buildPreviewCommand(d)
	cmdHeader := style.SectionTitle.Render("🚀 Execution Pipeline Preview:") + "\n"
	cmdContent := lipgloss.NewStyle().Width(contentWidth - 4).Render(cmdPreview)
	sb.WriteString(cmdBoxStyle.Width(contentWidth).Render(cmdHeader + cmdContent))
	sb.WriteString("\n")

	// 6. Action Menu & Hotkeys Dock (True 4-Column Grid Indentation)
	sb.WriteString(renderMenuDock(contentWidth, d.OpaqueBackdrop))

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

func renderMenuDock(width int, opaque bool) string {
	colW := (width - 6) / 3
	if colW < 24 {
		colW = 24
	}

	type dockItem struct {
		key   string
		label string
		sub   string
	}

	items := []dockItem{
		{"[Enter/1]", "Launch Game", "Run with current settings"},
		{"[2]", "Proton & Game Setup", "Runner, Exe, Quirks, ProtonDB"},
		{"[3]", "Performance & Sandbox", "Gamescope, CPU Cores, GPU, Xalia"},
		{"[4]", "Prefix & Overrides", "DLL Overrides, Reset, Health"},
		{"[5]", "Logs & Diagnostics", "Toggle Logging, Log Viewer"},
		{"[6]", "Settings & Help", "Backdrop, Manual, Quit"},
	}

	var row1, row2 []string
	for i, it := range items {
		rKey := style.KeyBadge.Render(it.key)
		title := lipgloss.NewStyle().Bold(true).Foreground(style.ColorPrimary).Render(it.label)
		sub := style.SubheaderStyle.Render(it.sub)

		card := fmt.Sprintf("%s %s\n    %s", rKey, title, sub)
		renderedCard := lipgloss.NewStyle().Width(colW).Render(card)

		if i < 3 {
			row1 = append(row1, renderedCard)
		} else {
			row2 = append(row2, renderedCard)
		}
	}

	r1 := lipgloss.JoinHorizontal(lipgloss.Top, row1...)
	r2 := lipgloss.JoinHorizontal(lipgloss.Top, row2...)
	grid := r1 + "\n\n" + r2

	dockContent := style.SectionTitle.Render("📂 Main Menu & Categories (Press key to open sub-menu):") + "\n\n" + grid

	return style.GetPanelStyle(opaque).Width(width).Render(dockContent)
}

