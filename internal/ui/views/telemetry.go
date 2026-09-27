package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/runner"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// TelemetryAction represents actions emitted from the Telemetry & Privacy Hub.
type TelemetryAction int

const (
	TelemetryActionNone TelemetryAction = iota
	TelemetryActionClose
	TelemetryActionToggle
	TelemetryActionWipe
)

// TelemetryView provides complete transparency into local session logging, privacy, and storage.
type TelemetryView struct {
	Config        *config.GameConfig
	GameDir       string
	History       []runner.SessionRecord
	InspectMode   bool
	StatusMessage string
	Width         int
	Height        int
}

// NewTelemetryView creates a new interactive view for the Local Telemetry & Privacy Hub.
func NewTelemetryView(cfg *config.GameConfig, gameDir string, width, height int) *TelemetryView {
	history, _ := runner.ReadSessionHistory(gameDir)
	return &TelemetryView{
		Config:  cfg,
		GameDir: gameDir,
		History: history,
		Width:   width,
		Height:  height,
	}
}

// RefreshHistory reloads the local session history ledger from disk.
func (v *TelemetryView) RefreshHistory() {
	history, _ := runner.ReadSessionHistory(v.GameDir)
	v.History = history
}

// Update handles keyboard input inside the Telemetry & Privacy view.
func (v *TelemetryView) Update(msg tea.Msg) (TelemetryAction, tea.Cmd) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case " ", "t", "T":
			return TelemetryActionToggle, nil
		case "v", "V":
			v.InspectMode = !v.InspectMode
			return TelemetryActionNone, nil
		case "w", "W":
			return TelemetryActionWipe, nil
		case "esc", "q", "enter":
			return TelemetryActionClose, nil
		}
	}
	return TelemetryActionNone, nil
}

// View renders the dedicated full-screen transparency audit.
func (v *TelemetryView) View() string {
	contentWidth := v.Width - 10
	if contentWidth < 68 {
		contentWidth = 68
	}
	if contentWidth > 96 {
		contentWidth = 96
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render("🛡️ LOCAL SESSION TELEMETRY & PRIVACY AUDIT")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Complete transparency into offline compatibility history and stored metrics")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	// Status Banner
	statusBadge := style.BadgeMuted.Render("○ DISABLED (Default / Zero Persistent Tracking)")
	if v.Config != nil && v.Config.EnableLocalTelemetry {
		statusBadge = style.BadgeSuccess.Render("● ACTIVE (Local-Only Anonymous Logging)")
	}
	b.WriteString("  Status: " + statusBadge + "\n\n")

	if v.StatusMessage != "" {
		b.WriteString(style.BadgeWarning.Render("  ℹ "+v.StatusMessage) + "\n\n")
	}

	if v.InspectMode {
		// History Inspection Mode
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("📜 RECORDED LOCAL HISTORY ENTRIES:") + "\n\n")
		if len(v.History) == 0 {
			b.WriteString(lipgloss.NewStyle().Foreground(style.ColorMuted).Render("  No session history currently recorded on disk.\n\n"))
		} else {
			for i := len(v.History) - 1; i >= 0 && i >= len(v.History)-8; i-- {
				h := v.History[i]
				statusStr := "Clean Exit (Code 0)"
				if h.CrashDetected || h.ExitCode != 0 {
					statusStr = fmt.Sprintf("Exit Code %d (Crash)", h.ExitCode)
				}
				durMins := int(h.DurationSeconds / 60)
				b.WriteString(fmt.Sprintf("  • [%s] %s | %s | %dm | %s\n",
					h.Timestamp.Format("2006-01-02 15:04"),
					h.RunnerName,
					h.TargetExe,
					durMins,
					statusStr,
				))
			}
			b.WriteString("\n")
		}
	} else {
		// Transparency Sections
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("ℹ️ OVERVIEW & PURPOSE") + "\n")
		b.WriteString(lipgloss.NewStyle().Foreground(style.ColorText).Render(
			"  When enabled, rpt keeps a local log of game sessions on your computer.\n"+
				"  This provides empirical data to back up ProtonDB reports (e.g. '5 sessions\n"+
				"  over 12 hours with 0 crashes') and verifies which runner version works best.\n\n",
		))

		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("📁 EXACT STORAGE LOCATION ON YOUR DISK") + "\n")
		historyPath := runner.SanitizePath(runner.GetHistoryFilePath(v.GameDir))
		b.WriteString(fmt.Sprintf("  • Per-Game Ledger: %s\n", lipgloss.NewStyle().Foreground(style.ColorSecondary).Render(historyPath)))
		b.WriteString(lipgloss.NewStyle().Foreground(style.ColorMuted).Render(
			"  Stored in human-readable JSON format. You can inspect or delete it at any time.\n\n",
		))

		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("📊 WHAT IS RECORDED (ONLY IF ENABLED)") + "\n")
		metrics := []string{
			"[✓] Game Executable Basename & Steam AppID (e.g. Endfield.exe / 4732690)",
			"[✓] Session Duration & Exit Code (e.g. 1h 45m, Exit 0 - Clean)",
			"[✓] Runner Version (e.g. GE-Proton9-25, Proton Experimental)",
			"[✓] Applied Quirks (e.g. Gamescope 1440p, P-Core Pinning, DXVK flags)",
			"[✓] Anonymized Hardware Summary (GPU Model, Driver string, CPU threads)",
		}
		for _, m := range metrics {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(style.ColorText).Render(m) + "\n")
		}
		b.WriteString("\n")

		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorDanger).Render("🚫 WHAT IS NEVER RECORDED OR TRANSMITTED (ZERO LEAKS)") + "\n")
		guarantees := []string{
			"[✗] NO personal file paths or usernames (always sanitized to ~/)",
			"[✗] NO Steam credentials, passwords, or authentication tokens",
			"[✗] NO game save files, player telemetry, or memory dumps",
			"[✗] NO network telemetry or phoning home to any third-party server (100% offline)",
		}
		for _, g := range guarantees {
			b.WriteString("  " + lipgloss.NewStyle().Foreground(style.ColorMuted).Render(g) + "\n")
		}
		b.WriteString("\n")
	}

	b.WriteString(divider + "\n")

	inspectLabel := "[v] Inspect Stored History"
	if v.InspectMode {
		inspectLabel = "[v] Hide History & View Audit"
	}

	dock := fmt.Sprintf("%s    %s    %s    %s",
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[Space/t] Toggle Logging"),
		lipgloss.NewStyle().Foreground(style.ColorSecondary).Render(inspectLabel),
		lipgloss.NewStyle().Foreground(style.ColorDanger).Render("[w] Wipe History"),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc] Return to Menu"),
	)
	b.WriteString(dock)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorPrimary).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
