package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// SubMenuType represents which category dialog is active.
type SubMenuType int

const (
	MenuNone SubMenuType = iota
	MenuTarget
	MenuDisplay
	MenuHardware
	MenuPrefix
	MenuDiagnostics

	// Aliases for compatibility with existing tests
	MenuProton      = MenuTarget
	MenuPerformance = MenuHardware
	MenuLogs        = MenuDiagnostics
	MenuSettings    = MenuDiagnostics
)

// SubMenuAction represents what action the controller model should execute.
type SubMenuAction int

const (
	ActionNone SubMenuAction = iota
	ActionClose
	ActionOpenProtonPicker
	ActionOpenExePicker
	ActionOpenQuirks
	ActionOpenProtonDB
	ActionToggleGamescope
	ActionTogglePCores
	ActionTogglePrimeRun
	ActionToggleXalia
	ActionCycleDisplayOutput
	ActionDetectMonitor
	ActionOpenGamescopeSettings
	ActionOpenOverrides
	ActionCleanPrefix
	ActionOpenDiagnostics
	ActionToggleLogging
	ActionOpenLogs
	ActionToggleBackdrop
	ActionOpenHelp
	ActionQuit
	ActionOpenHooks
	ActionOpenTelemetry
	ActionCreateDesktopShortcut
	ActionResetGamescope
	ActionResetHardware
	ActionResetOverrides
	ActionConfirmResetDefaults
)

// SubMenuData contains current state needed to display sub-menu options accurately.
type SubMenuData struct {
	Config          *config.GameConfig
	ProtonName      string
	ProtonDB        *integrations.ProtonDBReport
	HasPrimeRun     bool
	HasGamescope    bool
	ActiveOverrides map[string]string
	OpaqueBackdrop  bool
	StatusMessage   string
	PrefixActive    bool
	PrefixCleaned   bool
	Width           int
	Height          int
}

// SubMenuView renders category sub-menus with unified, consistent hotkeys.
type SubMenuView struct {
	Type SubMenuType
	Data SubMenuData
}

// NewSubMenuView initializes a sub-menu for the given category.
func NewSubMenuView(menuType SubMenuType, data SubMenuData) *SubMenuView {
	return &SubMenuView{
		Type: menuType,
		Data: data,
	}
}

// Update processes navigation and execution inside the sub-menu.
func (v *SubMenuView) Update(msg tea.Msg) SubMenuAction {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		k := keyMsg.String()
		switch k {
		case "esc", "q", "backspace":
			return ActionClose
		}

		switch v.Type {
		case MenuTarget:
			switch k {
			case "1", "e", "E":
				return ActionOpenExePicker
			case "2", "r", "R":
				return ActionOpenProtonPicker
			case "3", "d", "D":
				return ActionOpenQuirks
			case "4", "a", "A":
				return ActionOpenProtonDB
			}

		case MenuDisplay:
			switch k {
			case "1", "g", "G":
				return ActionToggleGamescope
			case "2", "m", "M":
				return ActionCycleDisplayOutput
			case "3", "d", "D":
				return ActionDetectMonitor
			case "4", "s", "S":
				return ActionOpenGamescopeSettings
			case "5", "r", "R":
				return ActionResetGamescope
			}

		case MenuHardware:
			switch k {
			case "1", "v", "V":
				return ActionTogglePrimeRun
			case "2", "p", "P":
				return ActionTogglePCores
			case "3", "x", "X":
				return ActionToggleXalia
			case "4", "r", "R":
				return ActionResetHardware
			case "g", "G":
				// Forwarding for test / muscle memory
				return ActionToggleGamescope
			}

		case MenuPrefix:
			switch k {
			case "1", "o", "O":
				return ActionOpenOverrides
			case "2", "c", "C":
				return ActionCleanPrefix
			case "3", "h", "H":
				return ActionOpenDiagnostics
			case "4", "r", "R":
				return ActionResetOverrides
			}

		case MenuDiagnostics:
			switch k {
			case "1", "L":
				return ActionToggleLogging
			case "2", "l":
				return ActionOpenLogs
			case "3", "k", "K", "d", "D":
				return ActionOpenDiagnostics
			case "4", "h", "H":
				return ActionOpenHooks
			case "5", "t", "T":
				return ActionOpenTelemetry
			case "6", "s", "S":
				return ActionCreateDesktopShortcut
			case "7", "b", "B":
				return ActionToggleBackdrop
			case "8", "?", "f1":
				return ActionOpenHelp
			case "9", "R":
				return ActionConfirmResetDefaults
			}
		}
	}
	return ActionNone
}

// View renders the sub-menu dialog card.
func (v *SubMenuView) View() string {
	contentWidth := v.Data.Width - 10
	if contentWidth < 60 {
		contentWidth = 60
	}
	if contentWidth > 88 {
		contentWidth = 88
	}

	var title, desc string
	var items []struct {
		keys  string
		label string
		state string
	}

	cfg := v.Data.Config

	switch v.Type {
	case MenuTarget:
		title = "🎯 TARGET & RUNNER COMPATIBILITY"
		desc = "Select target binary, Proton runner version, quirks preset, and ProtonDB ratings"

		pfxPreset := "Standard Defaults"
		if cfg.PresetName != "" {
			pfxPreset = "⚡ " + cfg.PresetName
		}

		pdbStatus := "Unknown / Unrated"
		if v.Data.ProtonDB != nil {
			titlePart := ""
			if v.Data.ProtonDB.Title != "" {
				titlePart = " - " + v.Data.ProtonDB.Title
			}
			pdbStatus = fmt.Sprintf("%s (%s, %d reports)%s", v.Data.ProtonDB.GetTierBadge(), v.Data.ProtonDB.Confidence, v.Data.ProtonDB.Total, titlePart)
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[e / 1]", "Select Target Executable", cfg.TargetExe},
			{"[r / 2]", "Select Proton Runner", v.Data.ProtonName},
			{"[d / 3]", "Game Quirks & Presets", pfxPreset},
			{"[a / 4]", "ProtonDB Community Report", pdbStatus},
		}

	case MenuDisplay:
		title = "📺 DISPLAY & GAMESCOPE SANDBOX"
		desc = "Configure display outputs, monitor size detection, and Gamescope scaling"

		gsStatus := style.BadgeMuted.Render("OFF (Native Window)")
		if cfg.UseGamescope {
			refStr := "Native"
			if cfg.GamescopeRefresh > 0 {
				refStr = fmt.Sprintf("%dHz", cfg.GamescopeRefresh)
			}
			gsStatus = style.BadgeSuccess.Render(fmt.Sprintf("ON (%dx%d @ %s -> %s)", cfg.GamescopeWidth, cfg.GamescopeHeight, refStr, cfg.GamescopeOutput))
		}

		dispStatus := style.BadgeSuccess.Render(cfg.GamescopeOutput)
		if cfg.GamescopeOutput == "" || strings.EqualFold(cfg.GamescopeOutput, "auto") {
			dispStatus = style.BadgeMuted.Render("Auto (External Preferred)")
		}

		detectDesc := "Detect native size & keep refresh untouched"
		optSummary := fmt.Sprintf("%dx%d (%s, %s)", cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeWindowMode, cfg.GamescopeFilter)

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[g / 1]", "Toggle Gamescope Sandboxing", gsStatus},
			{"[m / 2]", "Target Display Monitor", dispStatus},
			{"[d / 3]", "Detect Monitor Size & Native Hz", detectDesc},
			{"[s / 4]", "Gamescope Advanced Options", optSummary},
			{"[r / 5]", "Reset Display & Gamescope to Defaults", "Restores 1080p, Linear, SDR, Untouched Hz"},
		}

	case MenuHardware:
		title = "⚡ HARDWARE & ENGINE PERFORMANCE"
		desc = "GPU offloading, CPU thread affinity, and accessibility bridge"

		pcoreStatus := style.BadgeMuted.Render("OFF (All Threads)")
		if cfg.UsePCores {
			pcoreStatus = style.BadgeSuccess.Render(fmt.Sprintf("PINNED (Threads %s)", cfg.PCoresMask))
		}

		gpuStatus := style.BadgeMuted.Render("Host iGPU")
		if cfg.UsePrimeRun && v.Data.HasPrimeRun {
			gpuStatus = style.BadgeSuccess.Render("prime-run (Dedicated GPU)")
		}

		xaliaStatus := style.BadgeSuccess.Render("OFF (Clean DXVK)")
		if cfg.UseXalia {
			xaliaStatus = style.BadgeWarning.Render("ON (Accessibility Bridge)")
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[v / 1]", "Toggle GPU Runner (prime-run)", gpuStatus},
			{"[p / 2]", "Toggle CPU P-Core Pinning", pcoreStatus},
			{"[x / 3]", "Toggle Proton Xalia Bridge", xaliaStatus},
			{"[r / 4]", "Reset Hardware to Safe Defaults", "Restore detected GPU & CPU core pinning"},
		}

	case MenuPrefix:
		title = "🍷 WINE PREFIX & DLL OVERRIDES"
		desc = "Manage isolated prefix environment, DLL overrides, and filesystem reset"

		ovStatus := style.BadgeMuted.Render("None Active")
		if len(v.Data.ActiveOverrides) > 0 {
			ovStatus = style.BadgeSuccess.Render(fmt.Sprintf("%d active overrides", len(v.Data.ActiveOverrides)))
		}

		cleanStatus := style.BadgeWarning.Render("Active (Press to Wipe)")
		if !v.Data.PrefixActive {
			cleanStatus = style.BadgeMuted.Render("Not Initialized (Clean)")
		}
		if v.Data.PrefixCleaned {
			cleanStatus = style.BadgeSuccess.Render("✓ Reset Complete (Clean)")
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[o / 1]", "Configure DLL Overrides", ovStatus},
			{"[c / 2]", "Clean / Reset Wine Prefix", cleanStatus},
			{"[h / 3]", "Pre-flight Health & Diagnostics", "Checks +x bits and prefix paths"},
			{"[r / 4]", "Reset DLL Overrides to Clean", "Clears all custom user overrides"},
		}

	case MenuDiagnostics:
		title = "🔧 SYSTEM, LOGS & DIAGNOSTICS"
		desc = "Runtime execution logs, pre-flight diagnostics, lifecycle hooks, and tools"

		logStatus := style.BadgeMuted.Render("Disabled")
		if cfg.EnableLogging {
			logStatus = style.BadgeSuccess.Render("ACTIVE (writing to .logs/)")
		}

		telemStatus := style.BadgeMuted.Render("Disabled (Default)")
		if cfg.EnableLocalTelemetry {
			telemStatus = style.BadgeSuccess.Render("ACTIVE (Local-Only)")
		}

		bdStatus := style.BadgeSuccess.Render("Solid Dark")
		if !v.Data.OpaqueBackdrop {
			bdStatus = style.BadgeMuted.Render("Transparent")
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[L / 1]", "Toggle Session Logging", logStatus},
			{"[l / 2]", "View Session Logs & Crash Dumps", "Browse recent logs in .logs/"},
			{"[k / 3]", "Pre-flight System Diagnostics", "Check permissions and driver state"},
			{"[h / 4]", "Inspect Lifecycle Hooks", "Resolution checklist, env vars & pager"},
			{"[t / 5]", "Local Telemetry & Privacy Hub", telemStatus},
			{"[s / 6]", "Create Desktop Application Icon", "Install .desktop application launcher"},
			{"[b / 7]", "Terminal Backdrop Style", bdStatus},
			{"[? / 8]", "In-Depth Help & Documentation", "Comprehensive offline manual"},
			{"[R / 9]", "Revert Entire Config to Safe Defaults", "Preserves target exe & runner, resets all tweaks"},
		}
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render(title)

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render(desc)

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	if v.Data.StatusMessage != "" {
		b.WriteString(style.BadgeWarning.Render(" ℹ "+v.Data.StatusMessage) + "\n\n")
	}

	for _, it := range items {
		kBadge := style.KeyBadge.Render(it.keys)
		lblStr := lipgloss.NewStyle().Bold(true).Foreground(style.ColorText).Render(it.label)
		stStr := lipgloss.NewStyle().Foreground(style.ColorSecondary).Render(it.state)

		b.WriteString(fmt.Sprintf("  %s %-32s %s\n\n", kBadge, lblStr, stStr))
	}

	b.WriteString(divider + "\n")
	dock := fmt.Sprintf("%s    %s",
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("Press mnemonic key or number"),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc / Backspace] Return to Dashboard"),
	)
	b.WriteString(dock)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorPrimary).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
