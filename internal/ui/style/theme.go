package style

import "github.com/charmbracelet/lipgloss"

var (
	// Palette (Tokyo Night / Catppuccin inspired)
	ColorPrimary   = lipgloss.Color("#7aa2f7") // Blue
	ColorSecondary = lipgloss.Color("#bb9af7") // Purple
	ColorSuccess   = lipgloss.Color("#9ece6a") // Green
	ColorWarning   = lipgloss.Color("#e0af68") // Yellow
	ColorDanger    = lipgloss.Color("#f7768e") // Red
	ColorMuted     = lipgloss.Color("#565f89") // Gray
	ColorHighlight = lipgloss.Color("#7dcfff") // Cyan
	ColorBgDark    = lipgloss.Color("#16161e") // Deep Opaque Background for maximum contrast
	ColorText      = lipgloss.Color("#c0caf5") // Light Text
	ColorBorder    = lipgloss.Color("#3b4261") // Subtle Border

	// Typography Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary)

	SectionTitle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorSecondary)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	SubheaderStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	LabelStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorMuted)

	ValueStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Status Badges & Pills
	BadgeSuccess = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(ColorSuccess).
			Padding(0, 1)

	BadgeWarning = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(ColorWarning).
			Padding(0, 1)

	BadgeDanger = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#ffffff")).
			Background(ColorDanger).
			Padding(0, 1)

	BadgeMuted = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Padding(0, 1)

	BadgeHighlight = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(ColorHighlight).
			Padding(0, 1)

	// Hotkeys & Navigation
	KeyStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	KeyBadge = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#1a1b26")).
			Background(ColorHighlight).
			Padding(0, 1)

	KeyDesc = lipgloss.NewStyle().
			Foreground(ColorText)

	KeyPill = lipgloss.NewStyle().
			MarginRight(2)
)

// GetPanelStyle returns the panel border style, optionally filling with a solid dark background.
func GetPanelStyle(opaque bool) lipgloss.Style {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Padding(0, 1)
	if opaque {
		s = s.Background(ColorBgDark)
	}
	return s
}

// GetCommandBoxStyle returns the command box style with or without solid dark background.
func GetCommandBoxStyle(opaque bool) lipgloss.Style {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorBorder).
		Foreground(ColorHighlight).
		Padding(0, 1)
	if opaque {
		s = s.Background(ColorBgDark)
	}
	return s
}

// GetHeaderBarStyle returns the header bar style with or without solid dark background.
func GetHeaderBarStyle(opaque bool) lipgloss.Style {
	s := lipgloss.NewStyle().
		Bold(true).
		Foreground(ColorHighlight).
		Border(lipgloss.RoundedBorder()).
		BorderForeground(ColorPrimary).
		Padding(0, 1)
	if opaque {
		s = s.Background(ColorBgDark)
	}
	return s
}
