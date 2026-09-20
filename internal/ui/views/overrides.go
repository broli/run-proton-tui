package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// OverridesView manages the Wine registry DLL overrides.
type OverridesView struct {
	Presets   []proton.DLLPreset
	Active    map[string]bool
	Cursor    int
	PrefixDir string
	ProtonBin string
}

// NewOverridesView initializes the DLL overrides view.
func NewOverridesView(prefixDir, protonBin string, activeReg map[string]string) *OverridesView {
	presets := proton.GetStandardDLLPresets()
	activeMap := make(map[string]bool)

	for _, p := range presets {
		if _, ok := activeReg[p.DLL]; ok {
			activeMap[p.DLL] = true
		}
	}

	return &OverridesView{
		Presets:   presets,
		Active:    activeMap,
		Cursor:    0,
		PrefixDir: prefixDir,
		ProtonBin: protonBin,
	}
}

// Update handles navigation and toggles in the overrides view.
func (o *OverridesView) Update(msg tea.Msg) (*OverridesView, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if o.Cursor > 0 {
				o.Cursor--
			}
		case "down", "j":
			if o.Cursor < len(o.Presets)-1 {
				o.Cursor++
			}
		case " ": // Space to toggle
			if len(o.Presets) > 0 {
				target := o.Presets[o.Cursor]
				current := o.Active[target.DLL]
				if current {
					delete(o.Active, target.DLL)
					_ = proton.SetRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL, "")
				} else {
					o.Active[target.DLL] = true
					_ = proton.SetRegistryOverride(o.PrefixDir, o.ProtonBin, target.DLL, target.Mode)
				}
			}
		case "c": // Clear all
			o.Active = make(map[string]bool)
			_ = proton.ClearRegistryOverrides(o.PrefixDir, o.ProtonBin)
		case "enter", "esc", "q":
			return o, true // done
		}
	}
	return o, false
}

// View renders the overrides checklist.
func (o *OverridesView) View() string {
	var sb strings.Builder
	sb.WriteString(style.TitleStyle.Render("🧩 Wine Registry DLL Overrides (HKCU\\Software\\Wine\\DllOverrides)"))
	sb.WriteString("\n\n")

	sb.WriteString("Toggle with [Space], Clear all with [c], Press [Enter] or [Esc] to Return\n\n")

	for i, p := range o.Presets {
		cursor := "  "
		lineStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == o.Cursor {
			cursor = "👉"
			lineStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		box := "[ ]"
		if o.Active[p.DLL] {
			box = style.BadgeSuccess.Render("[✓]")
		}

		sb.WriteString(fmt.Sprintf("%s %s %s (%s = %s)\n", cursor, box, lineStyle.Render(p.Name), p.DLL, p.Mode))
		sb.WriteString(fmt.Sprintf("       %s\n", style.SubheaderStyle.Render(p.Description)))
	}

	return sb.String()
}
