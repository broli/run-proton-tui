package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
)

// DiagnosticsView displays the pre-flight permissions and health report.
type DiagnosticsView struct {
	Report *diagnostics.HealthReport
}

// NewDiagnosticsView initializes the diagnostics screen.
func NewDiagnosticsView(report *diagnostics.HealthReport) *DiagnosticsView {
	return &DiagnosticsView{
		Report: report,
	}
}

// Update handles return keys.
func (d *DiagnosticsView) Update(msg tea.Msg) (*DiagnosticsView, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "enter", "esc", "q", " ":
			return d, true // return to dashboard
		}
	}
	return d, false
}

// View renders the health report.
func (d *DiagnosticsView) View() string {
	var sb strings.Builder
	sb.WriteString(style.TitleStyle.Render("System & Game Health Diagnostics (Press [Enter] or [Esc] to Return)"))
	sb.WriteString("\n\n")

	rep := d.Report

	// Target Executable
	if rep.TargetExeExists {
		sb.WriteString(style.BadgeSuccess.Render(" ✓ Target Executable Exists "))
	} else {
		sb.WriteString(style.BadgeDanger.Render(" ✗ Target Executable Not Found "))
	}
	sb.WriteString("\n\n")

	// Prefix Containment Guardrail
	if rep.InsidePrefixHazard {
		sb.WriteString(style.BadgeDanger.Render(" HAZARD: Game binary is located inside Wine prefix! "))
		sb.WriteString("\n  Prefix cleaning has been locked to prevent game data loss.\n\n")
	} else {
		sb.WriteString(style.BadgeSuccess.Render(" ✓ Directory Guardrail: Game is safely separated from Wine prefix "))
		sb.WriteString("\n\n")
	}

	// 4. File Ownership & Permissions
	if len(rep.UnownedFiles) > 0 {
		sb.WriteString(style.BadgeWarning.Render(" ! Unowned Files Detected: "))
		sb.WriteString("\n")
		for _, f := range rep.UnownedFiles {
			sb.WriteString(fmt.Sprintf("  • %s\n", f))
		}
		sb.WriteString("  Tip: Run 'sudo chown -R $USER:$USER .' to fix.\n\n")
	} else {
		sb.WriteString(style.BadgeSuccess.Render(" ✓ All game files owned by current user "))
		sb.WriteString("\n\n")
	}

	// 5. Writeability
	if rep.GameDirWritable {
		sb.WriteString(style.BadgeSuccess.Render(" ✓ Game directory is writable "))
	} else {
		sb.WriteString(style.BadgeDanger.Render(" ✗ Game directory is NOT writable! "))
	}
	sb.WriteString("\n\n")

	if rep.PrefixDirWritable {
		sb.WriteString(style.BadgeSuccess.Render(" ✓ Wine prefix directory is writable "))
	} else {
		sb.WriteString(style.BadgeDanger.Render(" ✗ Wine prefix directory is NOT writable! "))
	}
	sb.WriteString("\n")

	return sb.String()
}
