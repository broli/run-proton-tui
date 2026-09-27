package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// CleanConfirmView renders a safety confirmation popup modal before clearing the Wine prefix.
type CleanConfirmView struct {
	GameTitle  string
	PrefixPath string
	BackupDir  string
	Width      int
	Height     int
}

// NewCleanConfirmView initializes the wine prefix reset confirmation dialog.
func NewCleanConfirmView(gameTitle, prefixPath, backupDir string, width, height int) *CleanConfirmView {
	return &CleanConfirmView{
		GameTitle:  gameTitle,
		PrefixPath: prefixPath,
		BackupDir:  backupDir,
		Width:      width,
		Height:     height,
	}
}

// Update processes confirmation keys.
// Returns confirmed (true if user pressed 'y' or 'enter' on confirm), and done (true when dialog should close).
func (v *CleanConfirmView) Update(msg tea.Msg) (confirmed bool, done bool) {
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
func (v *CleanConfirmView) View() string {
	contentWidth := v.Width - 14
	if contentWidth < 55 {
		contentWidth = 55
	}
	if contentWidth > 75 {
		contentWidth = 75
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorDanger).
		Render("⚠️  RESET WINE PREFIX CONFIRMATION")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("This operation will remove and recreate the isolated Wine container.")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	warnBox := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorWarning).
		Render("Are you sure you want to clean and reset this Wine prefix?")
	b.WriteString(warnBox + "\n\n")

	details := []struct {
		Label string
		Value string
	}{
		{"Game Title:", v.GameTitle},
		{"Target Prefix:", v.PrefixPath},
		{"Protection:", "SafeClean Auto-Preserve"},
	}

	for _, d := range details {
		b.WriteString(fmt.Sprintf("  • %-16s %s\n",
			lipgloss.NewStyle().Foreground(style.ColorMuted).Render(d.Label),
			lipgloss.NewStyle().Foreground(style.ColorHighlight).Render(d.Value),
		))
	}
	b.WriteString("\n")

	backupNotice := lipgloss.NewStyle().
		Foreground(style.ColorText).
		Background(style.ColorBgDark).
		Padding(0, 1).
		Width(contentWidth).
		Render("🛡️ SafeClean Protection: Your save files, AppData, and registry overrides will be backed up automatically before wiping.")
	b.WriteString(backupNotice + "\n\n")

	b.WriteString(divider + "\n")

	dock := fmt.Sprintf("%s    %s",
		lipgloss.NewStyle().Bold(true).Foreground(style.ColorDanger).Render("[y / Enter] Yes, Reset Prefix"),
		lipgloss.NewStyle().Bold(true).Foreground(style.ColorSuccess).Render("[n / Esc] Cancel & Keep Prefix"),
	)
	b.WriteString(dock)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorDanger).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
