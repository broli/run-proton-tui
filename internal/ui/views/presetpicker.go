package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/quirks"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// PresetPickerView displays detected UMU and curated quirks for user review before applying.
type PresetPickerView struct {
	Preset        *quirks.Preset
	GameTitle     string
	CurrentConfig *config.GameConfig
	Width         int
	Height        int
}

// NewPresetPickerView creates a new review dialog for presets.
func NewPresetPickerView(preset *quirks.Preset, gameTitle string, cfg *config.GameConfig, width, height int) *PresetPickerView {
	return &PresetPickerView{
		Preset:        preset,
		GameTitle:     gameTitle,
		CurrentConfig: cfg,
		Width:         width,
		Height:        height,
	}
}

// Update handles key navigation within the preset picker.
func (v *PresetPickerView) Update(msg tea.Msg) (applied bool, cancel bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch keyMsg.String() {
		case "enter", "y", "Y":
			if v.Preset != nil {
				return true, false
			}
			return false, true
		case "esc", "q", "n", "N":
			return false, true
		}
	}
	return false, false
}

// View renders the preset review dialog.
func (v *PresetPickerView) View() string {
	contentWidth := v.Width - 10
	if contentWidth < 60 {
		contentWidth = 60
	}
	if contentWidth > 90 {
		contentWidth = 90
	}

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render("⚡ GAME QUIRKS & PRESETS (Offline Engine)")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Review detected quirks and optimizations before applying them to .proton-config.toml")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	if v.Preset == nil {
		noMatchStyle := lipgloss.NewStyle().
			Foreground(style.ColorMuted).
			Italic(true)
		b.WriteString(noMatchStyle.Render("No specific non-standard quirks detected for this title.") + "\n\n")
		b.WriteString("The game will run cleanly using standard Proton & Wine settings.\n\n")
		b.WriteString(lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[Esc] Return to Dashboard"))
		return lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(style.ColorBorder).
			Padding(1, 2).
			Width(contentWidth + 4).
			Render(b.String())
	}

	p := v.Preset
	isActive := v.CurrentConfig != nil && (v.CurrentConfig.PresetName == p.Name || (p.UmuID != "" && v.CurrentConfig.UmuID == p.UmuID))

	activeBadge := ""
	if isActive {
		activeBadge = " " + lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#11111b")).
			Background(style.ColorSuccess).
			Padding(0, 1).
			Render("ACTIVE")
	}

	matchTitle := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorSuccess).
		Render("🎯 Matched Title: " + p.Name) + activeBadge

	matchSource := lipgloss.NewStyle().
		Foreground(style.ColorSecondary).
		Render("Source: " + p.MatchedSource)

	b.WriteString(matchTitle + "  " + matchSource + "\n\n")

	if p.SummaryNotes != "" {
		notesBox := lipgloss.NewStyle().
			Foreground(style.ColorText).
			Background(style.ColorBgDark).
			Padding(0, 1).
			Width(contentWidth).
			Render("ℹ️ " + p.SummaryNotes)
		b.WriteString(notesBox + "\n\n")
	}

	b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("Quirks & Runtime Configuration:") + "\n")

	if p.UmuID != "" {
		b.WriteString(fmt.Sprintf("  • %s: %s\n",
			lipgloss.NewStyle().Bold(true).Render("UMU GameFix ID"),
			lipgloss.NewStyle().Foreground(style.ColorHighlight).Render(p.UmuID),
		))
	}

	if len(p.EnvVars) > 0 {
		b.WriteString(fmt.Sprintf("  • %s:\n", lipgloss.NewStyle().Bold(true).Render("Custom Environment Variables")))
		for k, val := range p.EnvVars {
			b.WriteString(fmt.Sprintf("      %s = %s\n",
				lipgloss.NewStyle().Foreground(style.ColorMuted).Render(k),
				lipgloss.NewStyle().Foreground(style.ColorSecondary).Render(val),
			))
		}
	}

	if len(p.ExtraArgs) > 0 {
		b.WriteString(fmt.Sprintf("  • %s: %s\n",
			lipgloss.NewStyle().Bold(true).Render("Extra Launch Arguments"),
			lipgloss.NewStyle().Foreground(style.ColorHighlight).Render(strings.Join(p.ExtraArgs, " ")),
		))
	}

	if len(p.WaitProcesses) > 0 {
		b.WriteString(fmt.Sprintf("  • %s: %s\n",
			lipgloss.NewStyle().Bold(true).Render("Child Process Supervisor"),
			lipgloss.NewStyle().Foreground(style.ColorSecondary).Render(strings.Join(p.WaitProcesses, ", ")),
		))
	}

	if p.DisplayFile != "" {
		b.WriteString(fmt.Sprintf("  • %s: %s\n",
			lipgloss.NewStyle().Bold(true).Render("Display Socket Export"),
			lipgloss.NewStyle().Foreground(style.ColorMuted).Render(p.DisplayFile),
		))
	}

	if len(p.Profiles) > 0 {
		b.WriteString(fmt.Sprintf("  • %s:\n", lipgloss.NewStyle().Bold(true).Render("Dual Multi-Executable Profiles")))
		for exeKey := range p.Profiles {
			b.WriteString(fmt.Sprintf("      [%s] -> 2D Utility Mode (Host iGPU, No Gamescope)\n",
				lipgloss.NewStyle().Foreground(style.ColorHighlight).Render(exeKey),
			))
		}
	}

	b.WriteString("\n" + divider + "\n")
	enterLabel := "[Enter] Apply Presets & Save"
	if isActive {
		enterLabel = "[Enter] Re-apply Presets & Save"
	}
	dock := fmt.Sprintf("%s    %s",
		lipgloss.NewStyle().Bold(true).Foreground(style.ColorSuccess).Render(enterLabel),
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
