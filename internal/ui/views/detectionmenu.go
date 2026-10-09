package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/quirks"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// DetectionRoutine identifies a specific hardware or game discovery process.
type DetectionRoutine int

const (
	DetectDisplay DetectionRoutine = iota
	DetectGPU
	DetectCPU
	DetectQuirks
	DetectExesAndEmulators
	DetectRunners
	DetectAll
	routineCount
)

// DetectionResult holds the output and proposed actions of a completed detection routine.
type DetectionResult struct {
	Title           string
	Source          string
	Summary         string
	Findings        []string
	ApplyActionName string
	ApplyFn         func(cfg *config.GameConfig)
}

// DetectionMenuView provides an on-demand control panel allowing users to manually
// run any detection routine, view findings and source attribution, and choose whether to apply them.
type DetectionMenuView struct {
	Config        *config.GameConfig
	GameDir       string
	GameTitle     string
	Cursor        int
	ActiveResult  *DetectionResult
	StatusMessage string
	Width         int
	Height        int
}

// NewDetectionMenuView creates a new on-demand detection routines dialog.
func NewDetectionMenuView(cfg *config.GameConfig, gameDir, gameTitle string, width, height int) *DetectionMenuView {
	return &DetectionMenuView{
		Config:    cfg,
		GameDir:   gameDir,
		GameTitle: gameTitle,
		Cursor:    0,
		Width:     width,
		Height:    height,
	}
}

// Update handles navigation, execution, and confirmation within the detections menu.
func (v *DetectionMenuView) Update(msg tea.Msg) (done bool, applied bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		k := keyMsg.String()

		// If a result view is active, handle confirmation or returning to menu
		if v.ActiveResult != nil {
			switch k {
			case "enter", "y", "Y":
				if v.ActiveResult.ApplyFn != nil && v.Config != nil {
					v.ActiveResult.ApplyFn(v.Config)
					v.StatusMessage = fmt.Sprintf("✓ Applied findings from: %s", v.ActiveResult.Title)
					v.ActiveResult = nil
					return false, true
				}
				v.ActiveResult = nil
				return false, false
			case "esc", "q", "n", "N", "backspace":
				v.ActiveResult = nil
				return false, false
			}
			return false, false
		}

		// Menu navigation
		switch k {
		case "esc", "q":
			return true, false

		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
			} else {
				v.Cursor = int(routineCount) - 1
			}

		case "down", "j":
			if v.Cursor < int(routineCount)-1 {
				v.Cursor++
			} else {
				v.Cursor = 0
			}

		case "1":
			v.runRoutine(DetectDisplay)
		case "2":
			v.runRoutine(DetectGPU)
		case "3":
			v.runRoutine(DetectCPU)
		case "4":
			v.runRoutine(DetectQuirks)
		case "5":
			v.runRoutine(DetectExesAndEmulators)
		case "6":
			v.runRoutine(DetectRunners)
		case "7", "a", "A":
			v.runRoutine(DetectAll)

		case "enter", "space", " ":
			v.runRoutine(DetectionRoutine(v.Cursor))
		}
	}
	return false, false
}

func (v *DetectionMenuView) runRoutine(routine DetectionRoutine) {
	switch routine {
	case DetectDisplay:
		target := v.Config.GamescopeOutput
		w, h, err := hardware.DetectOutputResolution(target)
		outputs := hardware.GetConnectedDisplayOutputs()
		dispName := target
		if dispName == "" || strings.EqualFold(dispName, "auto") {
			dispName = "Auto-negotiated primary display"
		}
		status := "Successfully queried DRM connector modes"
		if err != nil {
			status = "Query fallback using default safe geometry"
		}

		v.ActiveResult = &DetectionResult{
			Title:   "Display & Monitor Resolution Probe",
			Source:  "DRM Sysfs Connector Query (/sys/class/drm)",
			Summary: status,
			Findings: []string{
				fmt.Sprintf("Active Display Target: %s", dispName),
				fmt.Sprintf("Connected Connectors:  %s", strings.Join(outputs, ", ")),
				fmt.Sprintf("Detected Resolution:   %dx%d (Native)", w, h),
				"Refresh Override:      0 Hz (Native / Host Negotiated)",
			},
			ApplyActionName: fmt.Sprintf("Set canvas to %dx%d (with '# auto detected from TUI' comment)", w, h),
			ApplyFn: func(cfg *config.GameConfig) {
				cfg.GamescopeWidth = w
				cfg.GamescopeHeight = h
				cfg.GamescopeRefresh = 0
				cfg.GeometryAutoDetected = true
			},
		}

	case DetectGPU:
		gpu, _ := hardware.DetectGPU()
		hasPrime := gpu != nil && gpu.HasPrimeRun
		var vendors []string
		if gpu != nil {
			if gpu.HasNvidia {
				vendors = append(vendors, "NVIDIA (dGPU)")
			}
			if gpu.HasAMD {
				vendors = append(vendors, "AMD")
			}
			if gpu.HasIntel {
				vendors = append(vendors, "Intel")
			}
		}
		vendorStr := "Unknown"
		if len(vendors) > 0 {
			vendorStr = strings.Join(vendors, " / ")
		}
		findings := []string{
			fmt.Sprintf("Primary GPU:        %s", vendorStr),
			fmt.Sprintf("prime-run Wrapper:  %v", hasPrime),
		}
		if gpu != nil && len(vendors) > 1 {
			findings = append(findings, "Architecture:       Hybrid Multi-GPU (iGPU + dGPU)")
		}

		v.ActiveResult = &DetectionResult{
			Title:   "GPU & Offload Probe",
			Source:  "PCI Bus & Vulkan ICD Device Inspection",
			Summary: "Analyzed graphics adapters and offload capabilities",
			Findings: findings,
			ApplyActionName: fmt.Sprintf("Set prime-run offload = %v", hasPrime),
			ApplyFn: func(cfg *config.GameConfig) {
				cfg.UsePrimeRun = hasPrime
			},
		}

	case DetectCPU:
		topo, err := hardware.DetectCPUTopology()
		isHybrid := topo != nil && topo.IsHybrid
		mask := ""
		if topo != nil {
			mask = topo.PCoresMask
		}
		summary := "Queried Linux CPU topology via sysfs"
		if err != nil {
			summary = "CPU topology query returned standard symmetric layout"
		}

		v.ActiveResult = &DetectionResult{
			Title:   "CPU Architecture & Core Topology Probe",
			Source:  "Linux Kernel CPU Sysfs (/sys/devices/system/cpu)",
			Summary: summary,
			Findings: []string{
				fmt.Sprintf("Hybrid Architecture: %v", isHybrid),
				fmt.Sprintf("P-Cores Thread Mask: %s", mask),
			},
			ApplyActionName: fmt.Sprintf("Configure CPU Pinning (use_pcores=%v, mask=%q)", isHybrid, mask),
			ApplyFn: func(cfg *config.GameConfig) {
				cfg.UsePCores = isHybrid
				cfg.PCoresMask = mask
			},
		}

	case DetectQuirks:
		preset := quirks.DetectQuirks(v.GameDir, v.Config.TargetExe, v.Config.AppID, v.Config.ProtonPath)
		if preset == nil {
			v.ActiveResult = &DetectionResult{
				Title:   "Game Quirks & Upstream Recipes Scan",
				Source:  "Local Manifests & Upstream UMU Database",
				Summary: "No specific quirks found for this game title.",
				Findings: []string{
					"Game runs safely with pure upstream Proton defaults (Clean Zero).",
				},
				ApplyActionName: "Keep Clean Zero defaults",
				ApplyFn:         nil,
			}
			return
		}

		var f []string
		f = append(f, fmt.Sprintf("Matched Game Title: %s", preset.Name))
		f = append(f, fmt.Sprintf("Rule Source:        %s", preset.MatchedSource))
		if preset.UmuID != "" {
			f = append(f, fmt.Sprintf("UMU GameFix ID:     %s", preset.UmuID))
		}
		for k, val := range preset.EnvVars {
			f = append(f, fmt.Sprintf("Env Var:            %s = %s", k, val))
		}
		if len(preset.ExtraArgs) > 0 {
			f = append(f, fmt.Sprintf("Extra Launch Args:  %s", strings.Join(preset.ExtraArgs, " ")))
		}

		v.ActiveResult = &DetectionResult{
			Title:           "Game Quirks & Upstream Recipes Scan",
			Source:          preset.MatchedSource,
			Summary:         preset.SummaryNotes,
			Findings:        f,
			ApplyActionName: fmt.Sprintf("Apply %s quirks preset to configuration", preset.Name),
			ApplyFn: func(cfg *config.GameConfig) {
				preset.ApplyToConfig(cfg, true)
			},
		}

	case DetectExesAndEmulators:
		exes := DiscoverExecutables(v.GameDir)
		emu, _ := integrations.ScanEmulators(v.GameDir)

		var f []string
		f = append(f, fmt.Sprintf("Discovered Windows Binaries: %d found", len(exes)))
		if len(exes) > 0 {
			f = append(f, fmt.Sprintf("Primary Target Candidate:    %s", exes[0].RelativePath))
		}
		if emu != nil && emu.Detected {
			f = append(f, fmt.Sprintf("Steam Emulator:              %s (AppID: %s)", emu.Type, emu.AppID))
		} else {
			f = append(f, "Steam Emulator:              None detected (standard retail binary)")
		}

		v.ActiveResult = &DetectionResult{
			Title:   "Executable & Steam Emulator Scan",
			Source:  "Game Directory PE Binary & DLL Analysis",
			Summary: "Scanned folder for launchable binaries and emulator configurations",
			Findings: f,
			ApplyActionName: "Adopt primary executable and detected AppID",
			ApplyFn: func(cfg *config.GameConfig) {
				if len(exes) > 0 && cfg.TargetExe == "" {
					cfg.TargetExe = exes[0].RelativePath
				}
				if emu != nil && emu.Detected && (cfg.AppID == "" || cfg.AppID == "0") {
					cfg.AppID = emu.AppID
				}
			},
		}

	case DetectRunners:
		runners, _ := proton.DiscoverRunners()
		var f []string
		f = append(f, fmt.Sprintf("Installed Proton Runners: %d found", len(runners)))
		for i, r := range runners {
			if i < 5 {
				f = append(f, fmt.Sprintf("  • [%s] %s (%s)", r.Type, r.Name, r.Path))
			}
		}
		if len(runners) > 5 {
			f = append(f, fmt.Sprintf("  ... and %d more runners", len(runners)-5))
		}

		v.ActiveResult = &DetectionResult{
			Title:   "Proton & Compatibility Tools Scan",
			Source:  "Steam compatibilitytools.d & System Paths",
			Summary: "Discovered installed Proton runners and Wine builds",
			Findings: f,
			ApplyActionName: "Set newest/preferred runner path if unset",
			ApplyFn: func(cfg *config.GameConfig) {
				if cfg.ProtonPath == "" && len(runners) > 0 {
					cfg.ProtonPath = runners[0].Path
				}
			},
		}

	case DetectAll:
		// Sweep through all detections
		gpu, _ := hardware.DetectGPU()
		topo, _ := hardware.DetectCPUTopology()
		w, h, _ := hardware.DetectOutputResolution(v.Config.GamescopeOutput)
		preset := quirks.DetectQuirks(v.GameDir, v.Config.TargetExe, v.Config.AppID, v.Config.ProtonPath)
		exes := DiscoverExecutables(v.GameDir)

		var f []string
		f = append(f, fmt.Sprintf("• Display: %dx%d (DRM native)", w, h))
		hasPrime := gpu != nil && gpu.HasPrimeRun
		f = append(f, fmt.Sprintf("• GPU:     prime-run capable = %v", hasPrime))
		isHybrid := topo != nil && topo.IsHybrid
		f = append(f, fmt.Sprintf("• CPU:     Intel Hybrid P-Cores = %v (Mask: %s)", isHybrid, topo.PCoresMask))
		if preset != nil {
			f = append(f, fmt.Sprintf("• Quirks:  %s (%s)", preset.Name, preset.MatchedSource))
		} else {
			f = append(f, "• Quirks:  Clean Zero (No game-specific quirks needed)")
		}
		if len(exes) > 0 {
			f = append(f, fmt.Sprintf("• Binary:  %s", exes[0].RelativePath))
		}

		v.ActiveResult = &DetectionResult{
			Title:   "Full Diagnostic Hardware & Game Sweep",
			Source:  "Unified System & Game Heuristics Suite",
			Summary: "Completed comprehensive probe across display, GPU, CPU, and game quirks",
			Findings: f,
			ApplyActionName: "Apply all recommended settings to active profile",
			ApplyFn: func(cfg *config.GameConfig) {
				cfg.GamescopeWidth = w
				cfg.GamescopeHeight = h
				cfg.GeometryAutoDetected = true
				cfg.UsePrimeRun = hasPrime
				if isHybrid {
					cfg.UsePCores = true
					cfg.PCoresMask = topo.PCoresMask
				}
				if preset != nil {
					preset.ApplyToConfig(cfg, true)
				}
			},
		}
	}
}

// View renders either the detection menu or the active routine results dialog.
func (v *DetectionMenuView) View() string {
	contentWidth := style.ClampWidth(v.Width, 10, 68, 92)

	if v.ActiveResult != nil {
		return v.renderResultDialog(contentWidth)
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render("🔍 HARDWARE & GAME DETECTION ROUTINES")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Run on-demand detection routines to probe hardware, displays, and game quirks")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	if v.StatusMessage != "" {
		b.WriteString(style.BadgeSuccess.Render(" ℹ "+v.StatusMessage) + "\n\n")
	}

	routines := []struct {
		key  string
		name string
		desc string
	}{
		{"1", "Detect Displays & Resolution", "Query DRM sysfs connectors for monitor resolution & refresh"},
		{"2", "Detect GPU & Prime-Run", "Probe PCI bus & Vulkan ICDs for NVIDIA dedicated GPU offload"},
		{"3", "Detect CPU Topology & Cores", "Check Linux sysfs for hybrid Intel P-cores and taskset mask"},
		{"4", "Scan for Game Quirks & Recipes", "Check local manifests, curated rules, and UMU ProtonFixes DB"},
		{"5", "Discover Executables & Emulators", "Scan game directory for PE binaries, launchers, and Steam emulators"},
		{"6", "Scan Installed Proton Runners", "Discover Proton-GE, CachyOS, DW-Proton, and Steam runners"},
		{"A", "Run All Detections & Diagnostics", "Execute full sweep across display, GPU, CPU, and game rules"},
	}

	for i, r := range routines {
		prefix := "   "
		lblStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		keyStyle := lipgloss.NewStyle().Bold(true).Foreground(style.ColorSecondary)
		if i == v.Cursor {
			prefix = " > "
			lblStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
			keyStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		b.WriteString(fmt.Sprintf("%s[%s] %-32s : %s\n", prefix, keyStyle.Render(r.key), lblStyle.Render(r.name), r.desc))
	}

	b.WriteString("\n" + divider + "\n")
	navHelp := fmt.Sprintf("%s    %s    %s",
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[↑/↓/1-6/A] Select Routine"),
		lipgloss.NewStyle().Foreground(style.ColorSuccess).Render("[Enter/Space] Run Selected"),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc/q] Return to Dashboard"),
	)
	b.WriteString(navHelp)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorPrimary).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}

func (v *DetectionMenuView) renderResultDialog(contentWidth int) string {
	res := v.ActiveResult
	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	title := lipgloss.NewStyle().Bold(true).Foreground(style.ColorPrimary).Render("🔍 " + res.Title)
	source := lipgloss.NewStyle().Foreground(style.ColorSecondary).Render("Source: " + res.Source)

	b.WriteString(title + "\n")
	b.WriteString(source + "\n")
	b.WriteString(divider + "\n\n")

	b.WriteString(lipgloss.NewStyle().Foreground(style.ColorText).Render(res.Summary) + "\n\n")

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("Findings & Data Collected:") + "\n")
	for _, f := range res.Findings {
		b.WriteString(fmt.Sprintf("  • %s\n", f))
	}
	b.WriteString("\n")

	if res.ApplyFn != nil {
		actionPrompt := lipgloss.NewStyle().Bold(true).Foreground(style.ColorSuccess).Render("Proposed Action: ") + res.ApplyActionName
		b.WriteString(actionPrompt + "\n\n")
		b.WriteString(divider + "\n")
		dock := fmt.Sprintf("%s    %s",
			lipgloss.NewStyle().Bold(true).Foreground(style.ColorSuccess).Render("[Y / Enter] Apply Finding to Profile"),
			lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[N / Esc] Cancel & Back to Menu"),
		)
		b.WriteString(dock)
	} else {
		b.WriteString(divider + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc / Enter] Back to Detection Menu"))
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorSuccess).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
