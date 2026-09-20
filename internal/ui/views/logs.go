package views

import (
	"fmt"
	"os"
	"strings"

	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// LogsView handles log selection and scrollable viewing.
type LogsView struct {
	Logs           []diagnostics.LogFileInfo
	Cursor         int
	ViewingContent bool
	Viewport       viewport.Model
	ActiveLogPath  string
}

// NewLogsView initializes the logs view.
func NewLogsView(logs []diagnostics.LogFileInfo, width, height int) *LogsView {
	vp := viewport.New(width, height-6)
	return &LogsView{
		Logs:           logs,
		Cursor:         0,
		ViewingContent: false,
		Viewport:       vp,
		ActiveLogPath:  "",
	}
}

// Update handles log navigation and scrolling.
func (l *LogsView) Update(msg tea.Msg) (*LogsView, bool) {
	if l.ViewingContent {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q":
				l.ViewingContent = false
				return l, false
			}
		}
		var cmd tea.Cmd
		l.Viewport, cmd = l.Viewport.Update(msg)
		_ = cmd
		return l, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if l.Cursor > 0 {
				l.Cursor--
			}
		case "down", "j":
			if l.Cursor < len(l.Logs)-1 {
				l.Cursor++
			}
		case "enter":
			if len(l.Logs) > 0 {
				target := l.Logs[l.Cursor]
				if data, err := os.ReadFile(target.Path); err == nil {
					l.ActiveLogPath = target.RelativePath
					l.Viewport.SetContent(string(data))
					l.Viewport.GotoBottom()
					l.ViewingContent = true
				}
			}
		case "esc", "q":
			return l, true // return to dashboard
		}
	}
	return l, false
}

// View renders either the log list or the active log viewport.
func (l *LogsView) View() string {
	if l.ViewingContent {
		header := style.TitleStyle.Render(fmt.Sprintf("📜 Viewing: %s (Press [Esc] to Return to List)", l.ActiveLogPath))
		footer := style.SubheaderStyle.Render(fmt.Sprintf("Scroll: ↑/k, ↓/j, PgUp, PgDown • Progress: %3.f%%", l.Viewport.ScrollPercent()*100))
		return lipgloss.JoinVertical(lipgloss.Left, header, l.Viewport.View(), footer)
	}

	var sb strings.Builder
	sb.WriteString(style.TitleStyle.Render("📋 Recent Session Logs & Diagnostics (Press [Enter] to View, [Esc] to Return)"))
	sb.WriteString("\n\n")

	if len(l.Logs) == 0 {
		sb.WriteString(style.BadgeWarning.Render("No log files found in .logs/ or game directory yet."))
		sb.WriteString("\nTip: Enable Logging Mode in rpt to generate detailed DXVK & Proton session logs.\n")
		return sb.String()
	}

	for i, item := range l.Logs {
		cursor := "  "
		lineStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == l.Cursor {
			cursor = "👉"
			lineStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		sizeStr := fmt.Sprintf("%.1f KB", float64(item.Size)/1024)
		timeStr := item.ModTime.Format("2006-01-02 15:04:05")
		sb.WriteString(fmt.Sprintf("%s %s (%s, %s)\n", cursor, lineStyle.Render(item.RelativePath), sizeStr, timeStr))
	}

	return sb.String()
}
