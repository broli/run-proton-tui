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
	MenuProton
	MenuPerformance
	MenuPrefix
	MenuLogs
	MenuSettings
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
	ActionOpenOverrides
	ActionCleanPrefix
	ActionOpenDiagnostics
	ActionToggleLogging
	ActionOpenLogs
	ActionToggleBackdrop
	ActionOpenHelp
	ActionQuit
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
			if v.Type == MenuSettings && k == "q" {
				return ActionQuit
			}
			return ActionClose
		}

		switch v.Type {
		case MenuProton:
			switch k {
			case "1", "r", "R":
				return ActionOpenProtonPicker
			case "2", "e", "E":
				return ActionOpenExePicker
			case "3", "d", "D":
				return ActionOpenQuirks
			case "4", "a", "A":
				return ActionOpenProtonDB
			}

		case MenuPerformance:
			switch k {
			case "1", "g", "G":
				return ActionToggleGamescope
			case "2", "p", "P":
				return ActionTogglePCores
			case "3", "v", "V":
				return ActionTogglePrimeRun
			case "4", "x", "X":
				return ActionToggleXalia
			case "5", "m", "M":
				return ActionCycleDisplayOutput
			}

		case MenuPrefix:
			switch k {
			case "1", "o", "O":
				return ActionOpenOverrides
			case "2", "c", "C":
				return ActionCleanPrefix
			case "3", "h", "H":
				return ActionOpenDiagnostics
			}

		case MenuLogs:
			switch k {
			case "1", "L":
				return ActionToggleLogging
			case "2", "l":
				return ActionOpenLogs
			case "3", "h", "H":
				return ActionOpenDiagnostics
			}

		case MenuSettings:
			switch k {
			case "1", "b", "B":
				return ActionToggleBackdrop
			case "2", "?", "f1":
				return ActionOpenHelp
			case "3", "q", "Q":
				return ActionQuit
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
	if contentWidth > 85 {
		contentWidth = 85
	}

	var title, desc string
	var items []struct {
		keys  string
		label string
		state string
	}

	cfg := v.Data.Config

	switch v.Type {
	case MenuProton:
		title = "🎮 PROTON & GAME TARGETS"
		desc = "Select Proton runner version, target binary, and compatibility profiles"

		pfxPreset := "Standard Defaults"
		if cfg.PresetName != "" {
			pfxPreset = "⚡ " + cfg.PresetName
		}

		pdbStatus := "Unknown / Unrated"
		if v.Data.ProtonDB != nil {
			pdbStatus = fmt.Sprintf("%s (%s, %d reports)", v.Data.ProtonDB.GetTierBadge(), v.Data.ProtonDB.Confidence, v.Data.ProtonDB.Total)
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[r / 1]", "Select Proton Runner", v.Data.ProtonName},
			{"[e / 2]", "Select Target Executable", cfg.TargetExe},
			{"[d / 3]", "Game Quirks & Presets", pfxPreset},
			{"[a / 4]", "ProtonDB Community Report", pdbStatus},
		}

	case MenuPerformance:
		title = "⚡ PERFORMANCE & SANDBOX"
		desc = "Hardware execution flags, display sandboxing, and CPU thread pinning"

		gsStatus := style.BadgeMuted.Render("OFF (Native Window)")
		if cfg.UseGamescope {
			gsStatus = style.BadgeSuccess.Render(fmt.Sprintf("ON (1080p @ %dHz -> %s)", cfg.GamescopeRefresh, cfg.GamescopeOutput))
		}

		pcoreStatus := style.BadgeMuted.Render("OFF (All Threads)")
		if cfg.UsePCores {
			pcoreStatus = style.BadgeSuccess.Render(fmt.Sprintf("PINNED (Threads %s)", cfg.PCoresMask))
		}

		gpuStatus := style.BadgeMuted.Render("Host iGPU")
		if cfg.UsePrimeRun && v.Data.HasPrimeRun {
			gpuStatus = style.BadgeSuccess.Render("prime-run (NVIDIA RTX)")
		}

		xaliaStatus := style.BadgeSuccess.Render("OFF (Clean DXVK)")
		if cfg.UseXalia {
			xaliaStatus = style.BadgeWarning.Render("ON (Accessibility Bridge)")
		}

		dispStatus := style.BadgeSuccess.Render(cfg.GamescopeOutput)
		if cfg.GamescopeOutput == "" || strings.EqualFold(cfg.GamescopeOutput, "auto") {
			dispStatus = style.BadgeMuted.Render("Auto (External Preferred)")
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[g / 1]", "Toggle Gamescope Sandboxing", gsStatus},
			{"[p / 2]", "Toggle CPU P-Core Pinning", pcoreStatus},
			{"[v / 3]", "Toggle GPU Runner (prime-run)", gpuStatus},
			{"[x / 4]", "Toggle Proton Xalia Bridge", xaliaStatus},
			{"[m / 5]", "Cycle Target Display Output", dispStatus},
		}

	case MenuPrefix:
		title = "🍷 WINE PREFIX & COMPATIBILITY"
		desc = "Manage isolated prefix environment, DLL overrides, and filesystem health"

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
		}

	case MenuLogs:
		title = "📜 LOGS & DIAGNOSTICS"
		desc = "Manage runtime execution logs, DXVK dumps, and pre-flight checks"

		logStatus := style.BadgeMuted.Render("Disabled")
		if cfg.EnableLogging {
			logStatus = style.BadgeSuccess.Render("ACTIVE (writing to .logs/)")
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[L / 1]", "Toggle Session Logging", logStatus},
			{"[l / 2]", "View Session Logs & Crash Dumps", "Browse recent logs in .logs/"},
			{"[h / 3]", "Pre-flight System Diagnostics", "Check permissions and driver state"},
		}

	case MenuSettings:
		title = "⚙️ SETTINGS & DOCUMENTATION"
		desc = "Configure visual interface preferences and browse offline manual"

		bdStatus := style.BadgeSuccess.Render("Solid Dark")
		if !v.Data.OpaqueBackdrop {
			bdStatus = style.BadgeMuted.Render("Transparent")
		}

		items = []struct {
			keys  string
			label string
			state string
		}{
			{"[b / 1]", "Terminal Backdrop Style", bdStatus},
			{"[? / 2]", "In-Depth Help & Troubleshooting", "Comprehensive offline guide"},
			{"[q / 3]", "Quit rpt Launcher", "Exit back to terminal"},
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
