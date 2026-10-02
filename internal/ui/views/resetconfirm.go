package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ResetConfirmView renders a safety confirmation popup modal before reverting to safe defaults.
type ResetConfirmView struct {
	GameTitle  string
	TargetExe  string
	ProtonName string
	Width      int
	Height     int
}

// NewResetConfirmView initializes the safe defaults confirmation dialog.
func NewResetConfirmView(gameTitle, targetExe, protonName string, width, height int) *ResetConfirmView {
	return &ResetConfirmView{
		GameTitle:  gameTitle,
		TargetExe:  targetExe,
		ProtonName: protonName,
		Width:      width,
		Height:     height,
	}
}

// Update processes confirmation keys.
// Returns confirmed (true if user pressed 'y' or 'enter' on confirm), and done (true when dialog should close).
func (v *ResetConfirmView) Update(msg tea.Msg) (confirmed bool, done bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "y", "Y", "enter":
			return true, true
		case "n", "N", "esc", "q":
			return false, true
		}
	}
	return false, false
}

// View renders the warning dialog.
func (v *ResetConfirmView) View() string {
	contentWidth := style.ClampWidth(v.Width, 14, 55, 80)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorWarning).
		Render("⚠️  REVERT TO SAFE DEFAULTS CONFIRMATION")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Restores standard, rock-solid stable settings for this game.")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	warnBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorHighlight).
		Render("Revert complex settings to safe defaults?")
	b.WriteString(warnBox + "\n\n")

	details := []struct {
		Label string
		Value string
	}{
		{"Game Title:", v.GameTitle},
		{"Target Binary:", fmt.Sprintf("Preserved (%s)", v.TargetExe)},
		{"Proton Runner:", fmt.Sprintf("Preserved (%s)", v.ProtonName)},
		{"Display / Sandbox:", "Reset to Safe (1080p, Linear, Untouched Hz, SDR)"},
		{"Hardware & CPU:", "Reset to Auto / Detected Hardware defaults"},
		{"DLL Overrides:", "Cleared (Default Wine registry only)"},
	}

	for _, d := range details {
		b.WriteString(fmt.Sprintf("  • %-18s %s\n",
			lipgloss.NewStyle().Foreground(style.ColorMuted).Render(d.Label),
			lipgloss.NewStyle().Foreground(style.ColorText).Render(d.Value),
		))
	}
	b.WriteString("\n")

	protectNotice := lipgloss.NewStyle().
		Foreground(style.ColorText).
		Background(style.ColorBgDark).
		Padding(0, 1).
		Width(contentWidth).
		Render("🛡️ SafeReset Protection: Target executable, AppID, and selected Proton runner are kept intact.")
	b.WriteString(protectNotice + "\n\n")

	b.WriteString(divider + "\n")

	dock := fmt.Sprintf("%s    %s",
		lipgloss.NewStyle().Bold(true).Foreground(style.ColorWarning).Render("[y / Enter] Yes, Revert to Safe Defaults"),
		lipgloss.NewStyle().Bold(true).Foreground(style.ColorSuccess).Render("[n / Esc] Cancel & Keep Settings"),
	)
	b.WriteString(dock)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorWarning).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
