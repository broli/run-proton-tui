package views

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProtonPickerView handles interactive Proton runner selection.
type ProtonPickerView struct {
	Runners     []proton.Runner
	Cursor      int
	Selected    proton.Runner
	CurrentPath string
	Width       int
	Height      int
}

// NewProtonPickerView initializes the Proton runner picker.
func NewProtonPickerView(runners []proton.Runner, currentPath string) *ProtonPickerView {
	cursor := 0
	realCurrent, _ := filepath.EvalSymlinks(currentPath)

	for i, r := range runners {
		realR, _ := filepath.EvalSymlinks(r.Path)
		if r.Path == currentPath || (realCurrent != "" && realR == realCurrent) {
			cursor = i
			break
		}
	}

	var sel proton.Runner
	if len(runners) > 0 {
		sel = runners[cursor]
	}

	return &ProtonPickerView{
		Runners:     runners,
		Cursor:      cursor,
		Selected:    sel,
		CurrentPath: currentPath,
		Width:       80,
		Height:      24,
	}
}

// SetDimensions updates the viewport dimensions for responsive layout.
func (p *ProtonPickerView) SetDimensions(w, h int) {
	p.Width = w
	p.Height = h
}

// Update handles navigation keys in the Proton picker.
func (p *ProtonPickerView) Update(msg tea.Msg) (*ProtonPickerView, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.Cursor > 0 {
				p.Cursor--
			} else if len(p.Runners) > 0 {
				p.Cursor = len(p.Runners) - 1
			}
		case "down", "j":
			if p.Cursor < len(p.Runners)-1 {
				p.Cursor++
			} else {
				p.Cursor = 0
			}
		case "home", "g":
			p.Cursor = 0
		case "end", "G":
			if len(p.Runners) > 0 {
				p.Cursor = len(p.Runners) - 1
			}
		case "enter", " ":
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

// getBadge returns the stylized family badge for a runner.
func getBadge(r proton.Runner) string {
	switch r.Type {
	case proton.RunnerTypeGE:
		return style.BadgeSuccess.Render("GE-Proton")
	case proton.RunnerTypeCachyOS:
		return style.BadgeHighlight.Render("CachyOS")
	case proton.RunnerTypeDW:
		return style.BadgeSecondary.Render("DW-Proton")
	case proton.RunnerTypeExperimental:
		return style.BadgeWarning.Render("Experimental")
	case proton.RunnerTypeValve:
		return style.BadgePrimary.Render("Valve Proton")
	case proton.RunnerTypeWine:
		return style.BadgeMuted.Render("Wine")
	default:
		if r.IsGE {
			return style.BadgeSuccess.Render("GE-Proton")
		} else if r.IsCachy {
			return style.BadgeHighlight.Render("CachyOS")
		} else if r.IsExperimental {
			return style.BadgeWarning.Render("Experimental")
		}
		return style.BadgeMuted.Render("Custom")
	}
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

	realCurrent, _ := filepath.EvalSymlinks(p.CurrentPath)

	maxVisible := 14
	if p.Height > 18 {
		maxVisible = p.Height - 10
	}
	if maxVisible < 5 {
		maxVisible = 5
	}

	start := 0
	end := len(p.Runners)

	if len(p.Runners) > maxVisible {
		half := maxVisible / 2
		start = p.Cursor - half
		if start < 0 {
			start = 0
		}
		end = start + maxVisible
		if end > len(p.Runners) {
			end = len(p.Runners)
			start = end - maxVisible
			if start < 0 {
				start = 0
			}
		}
	}

	if start > 0 {
		sb.WriteString(style.SubheaderStyle.Render(fmt.Sprintf("   ▲ ... %d more runner(s) above\n", start)))
	}

	for i := start; i < end; i++ {
		r := p.Runners[i]
		cursor := "  "
		lineStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == p.Cursor {
			cursor = "👉"
			lineStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		badge := getBadge(r)
		badgeCol := lipgloss.NewStyle().Width(16).Render(badge)

		realR, _ := filepath.EvalSymlinks(r.Path)
		isActive := r.Path == p.CurrentPath || (realCurrent != "" && realR == realCurrent)
		activeTag := ""
		if isActive {
			activeTag = " " + lipgloss.NewStyle().Foreground(style.ColorSuccess).Bold(true).Render("✓ Active")
		}

		sb.WriteString(fmt.Sprintf("%s %s %s%s\n", cursor, badgeCol, lineStyle.Render(r.Name), activeTag))
	}

	if end < len(p.Runners) {
		sb.WriteString(style.SubheaderStyle.Render(fmt.Sprintf("   ▼ ... %d more runner(s) below\n", len(p.Runners)-end)))
	}

	// Persistent detail and recommendation preview box at the bottom
	sb.WriteString("\n")
	divWidth := p.Width - 4
	if divWidth < 60 {
		divWidth = 60
	}
	if divWidth > 86 {
		divWidth = 86
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(style.ColorBorder).Render(strings.Repeat("─", divWidth)))
	sb.WriteString("\n")

	if p.Cursor >= 0 && p.Cursor < len(p.Runners) {
		cur := p.Runners[p.Cursor]
		badge := getBadge(cur)

		realCur, _ := filepath.EvalSymlinks(cur.Path)
		isActive := cur.Path == p.CurrentPath || (realCurrent != "" && realCur == realCurrent)
		activeStatus := ""
		if isActive {
			activeStatus = " " + lipgloss.NewStyle().Foreground(style.ColorSuccess).Bold(true).Render("[Currently Configured]")
		}

		sb.WriteString(fmt.Sprintf(" 💡 %s  %s%s\n", style.HeaderStyle.Render(cur.Name), badge, activeStatus))
		if cur.Recommendation != "" {
			sb.WriteString(fmt.Sprintf("    %s\n", style.SubheaderStyle.Render(cur.Recommendation)))
		}
		sb.WriteString(fmt.Sprintf("    %s %s\n", style.LabelStyle.Render("Path:"), style.ValueStyle.Render(cur.Path)))
	}

	sb.WriteString(style.SubheaderStyle.Render(" [↑/k, ↓/j] Navigate • [Enter/Space] Select • [Esc/q] Cancel"))
	return sb.String()
}

