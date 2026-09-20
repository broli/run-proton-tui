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
	ColorBgDark    = lipgloss.Color("#1a1b26") // Dark Background
	ColorText      = lipgloss.Color("#c0caf5") // Light Text

	// Typography & Layout Styles
	TitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorPrimary).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	HeaderStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(ColorHighlight)

	SubheaderStyle = lipgloss.NewStyle().
			Foreground(ColorMuted).
			Italic(true)

	CardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorMuted).
			Padding(0, 1).
			MarginBottom(1)

	ActiveCardStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(ColorPrimary).
			Padding(0, 1).
			MarginBottom(1)

	// Status Badges
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

	KeyDescStyle = lipgloss.NewStyle().
			Foreground(ColorText)

	// Command Box
	CommandBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.NormalBorder()).
			BorderForeground(ColorMuted).
			Foreground(ColorHighlight).
			Padding(0, 1).
			MarginTop(1)
)
