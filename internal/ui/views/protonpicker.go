package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProtonPickerView handles interactive Proton runner selection.
type ProtonPickerView struct {
	Runners  []proton.Runner
	Cursor   int
	Selected proton.Runner
}

// NewProtonPickerView initializes the Proton runner picker.
func NewProtonPickerView(runners []proton.Runner, currentPath string) *ProtonPickerView {
	cursor := 0
	for i, r := range runners {
		if r.Path == currentPath {
			cursor = i
			break
		}
	}
	var sel proton.Runner
	if len(runners) > 0 {
		sel = runners[cursor]
	}
	return &ProtonPickerView{
		Runners:  runners,
		Cursor:   cursor,
		Selected: sel,
	}
}

// Update handles navigation keys in the Proton picker.
func (p *ProtonPickerView) Update(msg tea.Msg) (*ProtonPickerView, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.Cursor > 0 {
				p.Cursor--
			}
		case "down", "j":
			if p.Cursor < len(p.Runners)-1 {
				p.Cursor++
			}
		case "enter":
			if len(p.Runners) > 0 {
				p.Selected = p.Runners[p.Cursor]
				return p, true, false // selected
			}
		case "esc", "q":
			return p, false, true // cancel
		}
	}
	return p, false, false
}

// View renders the Proton runner list.
func (p *ProtonPickerView) View() string {
	var sb strings.Builder
	sb.WriteString(style.TitleStyle.Render("⚡ Select Proton Compatibility Tool (Press [Enter] to Select, [Esc] to Cancel)"))
	sb.WriteString("\n\n")

	if len(p.Runners) == 0 {
		sb.WriteString(style.BadgeWarning.Render("No Proton runners found in Steam or compatibility directories."))
		return sb.String()
	}

	for i, r := range p.Runners {
		cursor := "  "
		lineStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == p.Cursor {
			cursor = "👉"
			lineStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		badge := ""
		if r.IsGE {
			badge = style.BadgeSuccess.Render("GE-Proton")
		} else if r.IsCachy {
			badge = style.BadgeHighlight.Render("CachyOS")
		} else if r.IsExperimental {
			badge = style.BadgeWarning.Render("Experimental")
		}

		sb.WriteString(fmt.Sprintf("%s %s %s\n", cursor, lineStyle.Render(r.Name), badge))
		if r.Recommendation != "" && i == p.Cursor {
			sb.WriteString(fmt.Sprintf("     %s\n", style.SubheaderStyle.Render(r.Recommendation)))
		}
	}

	return sb.String()
}
