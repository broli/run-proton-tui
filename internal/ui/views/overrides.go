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

// OverrideItem represents a single DLL override entry in the TUI checklist.
type OverrideItem struct {
	Name        string
	DLL         string
	Mode        string
	Description string
	IsCustom    bool
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

// OverridesView manages the Wine registry DLL overrides with presets and custom user definitions.
type OverridesView struct {
	Items     []OverrideItem
	Active    map[string]bool
	Modes     map[string]string
	Cursor    int
	PrefixDir string
	ProtonBin string
	InputMode bool
	TextInput textinput.Model
	StatusMsg string
}

// NewOverridesView initializes the DLL overrides view with standard presets and saved/active custom DLLs.
func NewOverridesView(prefixDir, protonBin string, activeReg map[string]string, savedOverrides map[string]string) *OverridesView {
	presets := proton.GetStandardDLLPresets()
	var items []OverrideItem
	activeMap := make(map[string]bool)
	modeMap := make(map[string]string)
	presetDLLs := make(map[string]bool)

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
		items = append(items, OverrideItem{
			Name:        p.Name,
			DLL:         p.DLL,
			Mode:        mode,
			Description: p.Description,
			IsCustom:    false,
		})
	}

	seen := make(map[string]bool)
	// Incorporate custom overrides found in user.reg
	for dll, mode := range activeReg {
		if !presetDLLs[dll] && !seen[dll] && dll != "" {
			seen[dll] = true
			activeMap[dll] = true
			if mode == "" {
				mode = "native,builtin"
			}
			modeMap[dll] = mode
			items = append(items, OverrideItem{
				Name:        fmt.Sprintf("Custom: %s", dll),
				DLL:         dll,
				Mode:        mode,
				Description: "User-defined custom DLL override",
				IsCustom:    true,
			})
		}
	}

	// Incorporate custom overrides saved in .proton-config.toml
	for dll, mode := range savedOverrides {
		if !presetDLLs[dll] && !seen[dll] && dll != "" {
			seen[dll] = true
			activeMap[dll] = true
			if mode == "" {
				mode = "native,builtin"
			}
			modeMap[dll] = mode
			items = append(items, OverrideItem{
				Name:        fmt.Sprintf("Custom: %s", dll),
				DLL:         dll,
				Mode:        mode,
				Description: "User-defined custom DLL override",
				IsCustom:    true,
			})
		}
	}

	ti := textinput.New()
	ti.Placeholder = "DLL name (e.g. xinput1_3, d3d11, winmm, dsound)"
	ti.CharLimit = 32
	ti.Width = 40

	return &OverridesView{
		Items:     items,
		Active:    activeMap,
		Modes:     modeMap,
		Cursor:    0,
		PrefixDir: prefixDir,
		ProtonBin: protonBin,
		InputMode: false,
		TextInput: ti,
		StatusMsg: "",
	}
}

// Update handles navigation, toggles, mode cycling, and custom DLL input.
func (o *OverridesView) Update(msg tea.Msg) (*OverridesView, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		if o.InputMode {
			switch msg.String() {
			case "enter":
				raw := strings.TrimSpace(o.TextInput.Value())
				raw = strings.TrimSuffix(strings.ToLower(raw), ".dll")
				if raw != "" {
					// Check if already in the list
					foundIndex := -1
					for i, it := range o.Items {
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
						}
						o.Items = append(o.Items, newItem)
						o.Active[raw] = true
						o.Modes[raw] = "native,builtin"
						o.Cursor = len(o.Items) - 1
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
			if o.Cursor < len(o.Items)-1 {
				o.Cursor++
			}
		case " ": // Space to toggle active / inactive
			if len(o.Items) > 0 {
				target := o.Items[o.Cursor]
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
			if len(o.Items) > 0 {
				target := &o.Items[o.Cursor]
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
			o.InputMode = true
			o.TextInput.Reset()
			o.TextInput.Focus()
			o.StatusMsg = "Type DLL name (e.g. xinput1_3, d3d11) and press [Enter]"
		case "d", "x", "delete": // Delete custom DLL
			if len(o.Items) > 0 {
				target := o.Items[o.Cursor]
				if target.IsCustom {
					_ = proton.DeleteRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL)
					delete(o.Active, target.DLL)
					delete(o.Modes, target.DLL)
					o.Items = append(o.Items[:o.Cursor], o.Items[o.Cursor+1:]...)
					if o.Cursor >= len(o.Items) && o.Cursor > 0 {
						o.Cursor = len(o.Items) - 1
					}
					o.StatusMsg = fmt.Sprintf("Deleted custom DLL override: %s", target.DLL)
				} else {
					o.StatusMsg = "Presets cannot be deleted, but can be deactivated with [Space]"
				}
			}
		case "c": // Clear all
			o.Active = make(map[string]bool)
			_ = proton.ClearRegistryOverrides(o.PrefixDir, o.ProtonBin)
			o.StatusMsg = "All registry DLL overrides cleared."
		case "enter", "esc", "q":
			return o, true // done
		}
	}
	return o, false
}

// GetActiveOverrides returns the active map of DLL -> Mode.
func (o *OverridesView) GetActiveOverrides() map[string]string {
	result := make(map[string]string)
	for dll, active := range o.Active {
		if active {
			mode := o.Modes[dll]
			if mode == "" {
				mode = "native,builtin"
			}
			result[dll] = mode
		}
	}
	return result
}

// View renders the interactive DLL overrides checklist and custom input dialog.
func (o *OverridesView) View() string {
	var sb strings.Builder

	header := style.TitleStyle.Render("🧩 Wine Registry DLL Overrides (HKCU\\Software\\Wine\\DllOverrides)")
	sb.WriteString(header + "\n\n")

	// Help bar
	helpText := fmt.Sprintf("%s Toggle  •  %s Cycle Mode  •  %s Add Custom DLL  •  %s Delete Custom  •  %s Return",
		style.KeyBadge.Render("[Space]"),
		style.KeyBadge.Render("[m]"),
		style.KeyBadge.Render("[a]"),
		style.KeyBadge.Render("[d]"),
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

	// Overrides List
	for i, it := range o.Items {
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

	return sb.String()
}
