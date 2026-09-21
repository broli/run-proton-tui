package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const defaultPageSize = 6

// OverrideItem represents a single DLL override entry in the TUI checklist.
type OverrideItem struct {
	Name        string
	DLL         string
	Mode        string
	Description string
	IsCustom    bool
	IsSystem    bool
}

var availableModes = []string{
	"native,builtin",
	"native",
	"builtin,native",
	"builtin",
	"disabled",
}

func nextMode(current string) string {
	for i, m := range availableModes {
		if m == current {
			return availableModes[(i+1)%len(availableModes)]
		}
	}
	return "native,builtin"
}

// OverridesView manages the Wine registry DLL overrides with paginated lists,
// separation of Wine system defaults, and custom user definitions.
type OverridesView struct {
	UserItems   []OverrideItem
	SystemItems []OverrideItem
	ShowSystem  bool
	Active      map[string]bool
	Modes       map[string]string
	Cursor      int
	PageSize    int
	PrefixDir   string
	ProtonBin   string
	InputMode   bool
	TextInput   textinput.Model
	StatusMsg   string
}

// NewOverridesView initializes the DLL overrides view, strictly separating
// user-defined/preset overrides from Wine's internal system runtime defaults.
func NewOverridesView(prefixDir, protonBin string, activeReg map[string]string, savedOverrides map[string]string) *OverridesView {
	presets := proton.GetStandardDLLPresets()
	var userItems []OverrideItem
	var systemItems []OverrideItem
	activeMap := make(map[string]bool)
	modeMap := make(map[string]string)
	presetDLLs := make(map[string]bool)

	// 1. Curated User Presets
	for _, p := range presets {
		presetDLLs[p.DLL] = true
		mode := p.Mode
		if regMode, ok := activeReg[p.DLL]; ok && regMode != "" {
			activeMap[p.DLL] = true
			mode = regMode
		} else if cfgMode, ok := savedOverrides[p.DLL]; ok && cfgMode != "" {
			activeMap[p.DLL] = true
			mode = cfgMode
		}
		modeMap[p.DLL] = mode
		userItems = append(userItems, OverrideItem{
			Name:        p.Name,
			DLL:         p.DLL,
			Mode:        mode,
			Description: p.Description,
			IsCustom:    false,
			IsSystem:    false,
		})
	}

	seen := make(map[string]bool)

	// 2. Classify registry entries into user vs system
	for dll, mode := range activeReg {
		if dll == "" || presetDLLs[dll] || seen[dll] {
			continue
		}
		seen[dll] = true

		if proton.IsWineDefaultDLL(dll) {
			// Wine internal system default (msvc, ucrt, etc.)
			systemItems = append(systemItems, OverrideItem{
				Name:        dll,
				DLL:         dll,
				Mode:        mode,
				Description: "Wine internal system runtime stub",
				IsCustom:    false,
				IsSystem:    true,
			})
		} else {
			// User-defined custom override
			activeMap[dll] = true
			if mode == "" {
				mode = "native,builtin"
			}
			modeMap[dll] = mode
			userItems = append(userItems, OverrideItem{
				Name:        fmt.Sprintf("Custom: %s", dll),
				DLL:         dll,
				Mode:        mode,
				Description: "User-defined custom DLL override",
				IsCustom:    true,
				IsSystem:    false,
			})
		}
	}

	// 3. Incorporate custom overrides from .proton-config.toml
	for dll, mode := range savedOverrides {
		if dll == "" || presetDLLs[dll] || seen[dll] || proton.IsWineDefaultDLL(dll) {
			continue
		}
		seen[dll] = true
		activeMap[dll] = true
		if mode == "" {
			mode = "native,builtin"
		}
		modeMap[dll] = mode
		userItems = append(userItems, OverrideItem{
			Name:        fmt.Sprintf("Custom: %s", dll),
			DLL:         dll,
			Mode:        mode,
			Description: "User-defined custom DLL override",
			IsCustom:    true,
			IsSystem:    false,
		})
	}

	ti := textinput.New()
	ti.Placeholder = "DLL name (e.g. xinput1_3, d3d11, winmm, dsound)"
	ti.CharLimit = 32
	ti.Width = 40

	return &OverridesView{
		UserItems:   userItems,
		SystemItems: systemItems,
		ShowSystem:  false,
		Active:      activeMap,
		Modes:       modeMap,
		Cursor:      0,
		PageSize:    defaultPageSize,
		PrefixDir:   prefixDir,
		ProtonBin:   protonBin,
		InputMode:   false,
		TextInput:   ti,
		StatusMsg:   "",
	}
}

func (o *OverridesView) activeItems() []OverrideItem {
	if o.ShowSystem {
		return o.SystemItems
	}
	return o.UserItems
}

// Update handles navigation, pagination, toggles, mode cycling, and custom DLL input.
func (o *OverridesView) Update(msg tea.Msg) (*OverridesView, bool) {
	items := o.activeItems()

	switch msg := msg.(type) {
	case tea.KeyMsg:
		if o.InputMode {
			switch msg.String() {
			case "enter":
				raw := strings.TrimSpace(o.TextInput.Value())
				raw = strings.TrimSuffix(strings.ToLower(raw), ".dll")
				if raw != "" {
					// Check if already in UserItems
					foundIndex := -1
					for i, it := range o.UserItems {
						if strings.EqualFold(it.DLL, raw) {
							foundIndex = i
							break
						}
					}

					if foundIndex >= 0 {
						o.Cursor = foundIndex
						o.Active[raw] = true
						_ = proton.SetRegistryOverride(o.PrefixDir, o.ProtonBin, raw, o.Modes[raw])
						o.StatusMsg = fmt.Sprintf("Override '%s' was already present and has been enabled.", raw)
					} else {
						newItem := OverrideItem{
							Name:        fmt.Sprintf("Custom: %s", raw),
							DLL:         raw,
							Mode:        "native,builtin",
							Description: "User-defined custom DLL override",
							IsCustom:    true,
							IsSystem:    false,
						}
						o.UserItems = append(o.UserItems, newItem)
						o.Active[raw] = true
						o.Modes[raw] = "native,builtin"
						o.Cursor = len(o.UserItems) - 1
						_ = proton.SetRegistryOverride(o.PrefixDir, o.ProtonBin, raw, "native,builtin")
						o.StatusMsg = fmt.Sprintf("Added custom DLL override: %s (mode: native,builtin)", raw)
					}
				}
				o.InputMode = false
				o.TextInput.Blur()
				return o, false

			case "esc":
				o.InputMode = false
				o.TextInput.Blur()
				o.StatusMsg = "Custom DLL entry cancelled."
				return o, false

			default:
				var cmd tea.Cmd
				o.TextInput, cmd = o.TextInput.Update(msg)
				_ = cmd
				return o, false
			}
		}

		// Standard list navigation
		switch msg.String() {
		case "up", "k":
			if o.Cursor > 0 {
				o.Cursor--
			}
		case "down", "j":
			if o.Cursor < len(items)-1 {
				o.Cursor++
			}
		case "pgup", "b":
			o.Cursor -= o.PageSize
			if o.Cursor < 0 {
				o.Cursor = 0
			}
		case "pgdown", "f":
			o.Cursor += o.PageSize
			if o.Cursor >= len(items) {
				o.Cursor = len(items) - 1
				if o.Cursor < 0 {
					o.Cursor = 0
				}
			}
		case "s": // Toggle showing Wine system defaults
			o.ShowSystem = !o.ShowSystem
			o.Cursor = 0
			if o.ShowSystem {
				o.StatusMsg = fmt.Sprintf("Viewing %d Wine internal system defaults (Read-Only). Press [s] to return to User Overrides.", len(o.SystemItems))
			} else {
				o.StatusMsg = "Viewing User & Game DLL Overrides."
			}
		case " ": // Space to toggle active / inactive
			if !o.ShowSystem && len(o.UserItems) > 0 {
				target := o.UserItems[o.Cursor]
				current := o.Active[target.DLL]
				if current {
					delete(o.Active, target.DLL)
					_ = proton.DeleteRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL)
					o.StatusMsg = fmt.Sprintf("Deactivated override: %s", target.DLL)
				} else {
					o.Active[target.DLL] = true
					mode := o.Modes[target.DLL]
					if mode == "" {
						mode = "native,builtin"
					}
					_ = proton.SetRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL, mode)
					o.StatusMsg = fmt.Sprintf("Activated override: %s = %s", target.DLL, mode)
				}
			}
		case "m": // Cycle mode for selected DLL
			if !o.ShowSystem && len(o.UserItems) > 0 {
				target := &o.UserItems[o.Cursor]
				currentMode := o.Modes[target.DLL]
				if currentMode == "" {
					currentMode = target.Mode
				}
				newMode := nextMode(currentMode)
				target.Mode = newMode
				o.Modes[target.DLL] = newMode
				if o.Active[target.DLL] {
					_ = proton.SetRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL, newMode)
				}
				o.StatusMsg = fmt.Sprintf("Set %s mode to: %s", target.DLL, newMode)
			}
		case "a", "+": // Add custom DLL
			o.ShowSystem = false // Ensure we are on UserItems view
			o.InputMode = true
			o.TextInput.Reset()
			o.TextInput.Focus()
			o.StatusMsg = "Type DLL name (e.g. xinput1_3, d3d11) and press [Enter]"
		case "d", "x", "delete": // Delete custom DLL
			if !o.ShowSystem && len(o.UserItems) > 0 {
				target := o.UserItems[o.Cursor]
				if target.IsCustom {
					_ = proton.DeleteRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL)
					delete(o.Active, target.DLL)
					delete(o.Modes, target.DLL)
					o.UserItems = append(o.UserItems[:o.Cursor], o.UserItems[o.Cursor+1:]...)
					if o.Cursor >= len(o.UserItems) && o.Cursor > 0 {
						o.Cursor = len(o.UserItems) - 1
					}
					o.StatusMsg = fmt.Sprintf("Deleted custom DLL override: %s", target.DLL)
				} else {
					o.StatusMsg = "Presets cannot be deleted, but can be deactivated with [Space]"
				}
			}
		case "c": // Clear all user overrides
			if !o.ShowSystem {
				for _, it := range o.UserItems {
					delete(o.Active, it.DLL)
					_ = proton.DeleteRegistryOverride(o.PrefixDir, o.ProtonBin, it.DLL)
				}
				o.StatusMsg = "All user DLL overrides cleared."
			}
		case "enter", "esc", "q":
			return o, true // done
		}
	}
	return o, false
}

// GetActiveOverrides returns only active user-defined and preset overrides (never Wine system defaults).
func (o *OverridesView) GetActiveOverrides() map[string]string {
	result := make(map[string]string)
	for dll, active := range o.Active {
		if active && !proton.IsWineDefaultDLL(dll) {
			mode := o.Modes[dll]
			if mode == "" {
				mode = "native,builtin"
			}
			result[dll] = mode
		}
	}
	return result
}

// View renders the paginated DLL overrides checklist, tab switcher, and custom input dialog.
func (o *OverridesView) View() string {
	var sb strings.Builder

	header := style.TitleStyle.Render("🧩 Wine Registry DLL Overrides (HKCU\\Software\\Wine\\DllOverrides)")
	sb.WriteString(header + "\n\n")

	tabUser := style.BadgeHighlight.Render("[1] User & Game Overrides")
	tabSys := style.BadgeMuted.Render(fmt.Sprintf("[2] Wine System Defaults (%d)", len(o.SystemItems)))
	if o.ShowSystem {
		tabUser = style.BadgeMuted.Render("[1] User & Game Overrides")
		tabSys = style.BadgeHighlight.Render(fmt.Sprintf("[2] Wine System Defaults (%d)", len(o.SystemItems)))
	}
	sb.WriteString(fmt.Sprintf("Category: %s  %s  (Press [s] to switch)\n\n", tabUser, tabSys))

	// Help bar
	helpText := fmt.Sprintf("%s Toggle  •  %s Cycle Mode  •  %s Add DLL  •  %s Delete  •  %s Page  •  %s Return",
		style.KeyBadge.Render("[Space]"),
		style.KeyBadge.Render("[m]"),
		style.KeyBadge.Render("[a]"),
		style.KeyBadge.Render("[d]"),
		style.KeyBadge.Render("[PgUp/PgDn]"),
		style.KeyBadge.Render("[Enter/Esc]"),
	)
	sb.WriteString(helpText + "\n\n")

	// Status message if present
	if o.StatusMsg != "" {
		sb.WriteString(style.BadgeWarning.Render(" ℹ "+o.StatusMsg) + "\n\n")
	}

	// Custom DLL Input Dialog
	if o.InputMode {
		inputCard := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(style.ColorHighlight).
			Padding(1, 2).
			Render(fmt.Sprintf(
				"%s\n\n%s\n\n%s",
				style.HeaderStyle.Render("➕ Add Custom Wine DLL Override:"),
				o.TextInput.View(),
				style.SubheaderStyle.Render("Press [Enter] to add override, [Esc] to cancel"),
			))
		sb.WriteString(inputCard + "\n\n")
	}

	items := o.activeItems()
	if len(items) == 0 {
		if o.ShowSystem {
			sb.WriteString(style.SubheaderStyle.Render("  No Wine system default overrides found in this prefix.\n\n"))
		} else {
			sb.WriteString(style.SubheaderStyle.Render("  No user overrides defined yet. Press [a] to add your first custom DLL!\n\n"))
		}
		return sb.String()
	}

	// Calculate pagination
	totalPages := (len(items) + o.PageSize - 1) / o.PageSize
	if totalPages == 0 {
		totalPages = 1
	}
	currentPage := o.Cursor / o.PageSize
	startIndex := currentPage * o.PageSize
	endIndex := startIndex + o.PageSize
	if endIndex > len(items) {
		endIndex = len(items)
	}

	// Render current page items
	for i := startIndex; i < endIndex; i++ {
		it := items[i]
		cursor := "  "
		lineStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == o.Cursor {
			cursor = "👉"
			lineStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		box := style.BadgeMuted.Render("[ ]")
		if o.Active[it.DLL] {
			box = style.BadgeSuccess.Render("[✓]")
		}

		tag := style.BadgeHighlight.Render("PRESET")
		if it.IsCustom {
			tag = style.BadgeWarning.Render("CUSTOM")
		} else if it.IsSystem {
			tag = style.BadgeMuted.Render("SYSTEM")
		}

		mode := it.Mode
		if m, ok := o.Modes[it.DLL]; ok && m != "" {
			mode = m
		}

		modeBadge := lipgloss.NewStyle().
			Bold(true).
			Foreground(style.ColorSecondary).
			Render(fmt.Sprintf("[%s]", mode))

		sb.WriteString(fmt.Sprintf("%s %s %s %s %s\n",
			cursor,
			box,
			tag,
			lineStyle.Render(fmt.Sprintf("%-28s", it.Name+" ("+it.DLL+".dll)")),
			modeBadge,
		))
		sb.WriteString(fmt.Sprintf("          %s\n", style.SubheaderStyle.Render(it.Description)))
	}

	// Pagination footer
	sb.WriteString("\n")
	pagerFooter := style.SubheaderStyle.Render(fmt.Sprintf(
		"─── Page %d of %d (Showing %d-%d of %d items) • Use [↑/k, ↓/j, PgUp, PgDn] to navigate ───",
		currentPage+1,
		totalPages,
		startIndex+1,
		endIndex,
		len(items),
	))
	sb.WriteString(pagerFooter + "\n")

	return sb.String()
}
