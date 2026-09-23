package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ProtonDBView renders a full interactive report on ProtonDB community ratings.
type ProtonDBView struct {
	Report      *integrations.ProtonDBReport
	GameTitle   string
	AppID       string
	Status      string
	SearchMode  bool
	SearchInput textinput.Model
	Width       int
	Height      int
}

// NewProtonDBView creates a new view for displaying ProtonDB details.
func NewProtonDBView(report *integrations.ProtonDBReport, title, appID string, width, height int) *ProtonDBView {
	ti := textinput.New()
	ti.Placeholder = "Enter game title or Steam AppID (e.g. Endfield or 4732690)"
	ti.CharLimit = 64
	ti.Width = 55

	return &ProtonDBView{
		Report:      report,
		GameTitle:   title,
		AppID:       appID,
		SearchInput: ti,
		Width:       width,
		Height:      height,
	}
}

// Update handles navigation and custom search within the ProtonDB view.
func (v *ProtonDBView) Update(msg tea.Msg) (customQuery string, refresh bool, done bool, cmd tea.Cmd) {
	if v.SearchMode {
		if keyMsg, ok := msg.(tea.KeyMsg); ok {
			switch keyMsg.String() {
			case "enter":
				val := strings.TrimSpace(v.SearchInput.Value())
				v.SearchMode = false
				v.SearchInput.Blur()
				if val != "" {
					return val, true, false, nil
				}
				return "", false, false, nil
			case "esc":
				v.SearchMode = false
				v.SearchInput.Blur()
				return "", false, false, nil
			}
		}
		var inputCmd tea.Cmd
		v.SearchInput, inputCmd = v.SearchInput.Update(msg)
		return "", false, false, inputCmd
	}

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "r", "R":
			return "", true, false, nil
		case "s", "S", "/":
			v.SearchMode = true
			v.SearchInput.SetValue("")
			cmd := v.SearchInput.Focus()
			return "", false, false, cmd
		case "esc", "q", "enter":
			return "", false, true, nil
		}
	}
	return "", false, false, nil
}

// View renders the ProtonDB report card.
func (v *ProtonDBView) View() string {
	contentWidth := v.Width - 10
	if contentWidth < 60 {
		contentWidth = 60
	}
	if contentWidth > 85 {
		contentWidth = 85
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render("🌟 PROTONDB COMMUNITY REPORT")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Aggregated Linux compatibility reports from protondb.com")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	// Status or Error Banner
	if v.Status != "" {
		stStyle := lipgloss.NewStyle().
			Bold(true).
			Padding(0, 1).
			Width(contentWidth)
		if strings.Contains(strings.ToLower(v.Status), "error") || strings.Contains(strings.ToLower(v.Status), "fail") || strings.Contains(strings.ToLower(v.Status), "no steam title") {
			stStyle = stStyle.Foreground(style.ColorDanger).Background(style.ColorBgDark)
		} else {
			stStyle = stStyle.Foreground(style.ColorHighlight).Background(style.ColorBgDark)
		}
		b.WriteString(stStyle.Render("ℹ "+v.Status) + "\n\n")
	}

	if v.SearchMode {
		searchPrompt := lipgloss.NewStyle().
			Bold(true).
			Foreground(style.ColorHighlight).
			Render("🔍 Search Steam Store / Set AppID:")
		b.WriteString(searchPrompt + "\n")
		b.WriteString(v.SearchInput.View() + "\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Enter] Search Steam & ProtonDB    [Esc] Cancel") + "\n\n")
	}

	if v.Report == nil {
		if !v.SearchMode {
			b.WriteString(lipgloss.NewStyle().Foreground(style.ColorWarning).Render("No ProtonDB report loaded yet.") + "\n\n")
			b.WriteString(fmt.Sprintf("AppID: %s │ Game Title: %s\n\n", v.AppID, v.GameTitle))
			b.WriteString("Press [r] to query/retry, [s] to search by custom title or AppID, or [Esc] to return.\n\n")
		}
		b.WriteString(divider + "\n")
		dock := fmt.Sprintf("%s    %s    %s",
			lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[r] Retry Fetch"),
			lipgloss.NewStyle().Foreground(style.ColorSecondary).Render("[s] Search Custom Game / AppID"),
			lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc] Return to Dashboard"),
		)
		b.WriteString(dock)

		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(style.ColorBorder).
			Padding(1, 2).
			Width(contentWidth + 4).
			Render(b.String())
	}

	r := v.Report
	gameLine := fmt.Sprintf("%s %s  %s",
		lipgloss.NewStyle().Bold(true).Render("Game:"),
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render(v.GameTitle),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render(fmt.Sprintf("(AppID: %s)", v.AppID)),
	)
	b.WriteString(gameLine + "\n\n")

	// Overall Tier Banner
	tierBanner := fmt.Sprintf("Overall Compatibility: %s", r.GetTierBadge())
	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorText).Render(tierBanner) + "\n\n")

	// Metrics Grid
	metrics := []struct {
		Label string
		Value string
	}{
		{"Total Community Reports:", fmt.Sprintf("%d user reports", r.Total)},
		{"Community Confidence:", strings.Title(r.Confidence)},
		{"Trending Recent Tier:", strings.ToUpper(r.TrendingTier)},
		{"Best Reported Tier:", strings.ToUpper(r.BestReportedTier)},
	}

	for _, m := range metrics {
		b.WriteString(fmt.Sprintf("  • %-26s %s\n",
			lipgloss.NewStyle().Foreground(style.ColorMuted).Render(m.Label),
			lipgloss.NewStyle().Foreground(style.ColorSecondary).Render(m.Value),
		))
	}

	b.WriteString("\n")

	// Compatibility Guidance
	guidance := ""
	switch strings.ToLower(r.Tier) {
	case "platinum":
		guidance = "⭐ Platinum: Runs flawlessly out of the box with default Proton. No tweaks needed."
	case "gold":
		guidance = "🥇 Gold: Runs smoothly after minor tweaks (e.g. Proton-GE or built-in gamefixes)."
	case "silver":
		guidance = "🥈 Silver: Generally playable, but may encounter minor audio, video, or graphical glitches."
	case "bronze":
		guidance = "🥉 Bronze: Significant issues or frequent crashes under standard Wine/Proton."
	case "borked":
		guidance = "❌ Borked: Will not start or crashes immediately (often due to anti-cheat or media foundation)."
	default:
		guidance = "ℹ️ Compatibility status based on community crowd-sourced reports."
	}

	notesBox := lipgloss.NewStyle().
		Foreground(style.ColorText).
		Background(style.ColorBgDark).
		Padding(0, 1).
		Width(contentWidth).
		Render(guidance)
	b.WriteString(notesBox + "\n\n")

	b.WriteString(divider + "\n")
	dock := fmt.Sprintf("%s    %s    %s",
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[r] Refresh Report"),
		lipgloss.NewStyle().Foreground(style.ColorSecondary).Render("[s] Search Custom Game / AppID"),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc] Return to Dashboard"),
	)
	b.WriteString(dock)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorPrimary).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
