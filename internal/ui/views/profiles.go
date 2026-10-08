package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type profileMode int

const (
	profileModeList profileMode = iota
	profileModeNew
	profileModeConfirmDelete
)

// ProfilesView provides interactive management of game configuration profiles.
type ProfilesView struct {
	ConfigFile    *config.GameConfigFile
	Cursor        int
	Mode          profileMode
	NameInput     textinput.Model
	DeleteTarget  string
	StatusMessage string
	Width         int
	Height        int
}

// NewProfilesView creates a new profile management view.
func NewProfilesView(fileCfg *config.GameConfigFile, width, height int) *ProfilesView {
	ti := textinput.New()
	ti.Placeholder = "e.g. launcher, high-graphics, portable"
	ti.CharLimit = 32
	ti.Width = 36

	v := &ProfilesView{
		ConfigFile: fileCfg,
		Cursor:     0,
		Mode:       profileModeList,
		NameInput:  ti,
		Width:      width,
		Height:     height,
	}

	// Match cursor to active profile if possible
	names := fileCfg.ProfileNames()
	for i, name := range names {
		if name == fileCfg.ActiveProfile {
			v.Cursor = i
			break
		}
	}

	return v
}

// Update handles keypresses and actions inside the profile manager.
// Returns (activeProfileChanged bool, configModified bool, returnToDashboard bool)
func (v *ProfilesView) Update(msg tea.Msg) (bool, bool, bool) {
	names := v.ConfigFile.ProfileNames()

	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		switch v.Mode {
		case profileModeNew:
			switch keyMsg.String() {
			case "enter":
				newName := strings.TrimSpace(v.NameInput.Value())
				if newName == "" {
					v.StatusMessage = "⚠️ Profile name cannot be empty"
					return false, false, false
				}
				// Clone from currently selected profile
				srcName := v.ConfigFile.ActiveProfile
				if v.Cursor >= 0 && v.Cursor < len(names) {
					srcName = names[v.Cursor]
				}
				if err := v.ConfigFile.CloneProfile(srcName, newName); err != nil {
					v.StatusMessage = fmt.Sprintf("⚠️ %v", err)
					return false, false, false
				}
				_ = v.ConfigFile.SetActiveProfile(newName)
				v.Mode = profileModeList
				v.StatusMessage = fmt.Sprintf("✓ Created and switched to profile %q", newName)
				// Reposition cursor
				newNames := v.ConfigFile.ProfileNames()
				for i, n := range newNames {
					if n == newName {
						v.Cursor = i
						break
					}
				}
				return true, true, false

			case "esc":
				v.Mode = profileModeList
				v.StatusMessage = ""
				return false, false, false

			default:
				var cmd tea.Cmd
				v.NameInput, cmd = v.NameInput.Update(msg)
				_ = cmd
				return false, false, false
			}

		case profileModeConfirmDelete:
			switch keyMsg.String() {
			case "y", "Y", "enter":
				target := v.DeleteTarget
				if err := v.ConfigFile.DeleteProfile(target); err != nil {
					v.StatusMessage = fmt.Sprintf("⚠️ %v", err)
				} else {
					v.StatusMessage = fmt.Sprintf("✓ Deleted profile %q", target)
				}
				v.Mode = profileModeList
				if v.Cursor >= len(v.ConfigFile.ProfileNames()) {
					v.Cursor = max(0, len(v.ConfigFile.ProfileNames())-1)
				}
				return true, true, false

			case "n", "N", "esc", "q":
				v.Mode = profileModeList
				v.StatusMessage = "Deletion cancelled"
				return false, false, false
			}

		case profileModeList:
			switch keyMsg.String() {
			case "up", "k":
				if v.Cursor > 0 {
					v.Cursor--
				} else {
					v.Cursor = len(names) - 1
				}
				return false, false, false

			case "down", "j":
				if v.Cursor < len(names)-1 {
					v.Cursor++
				} else {
					v.Cursor = 0
				}
				return false, false, false

			case "enter":
				if v.Cursor >= 0 && v.Cursor < len(names) {
					selected := names[v.Cursor]
					_ = v.ConfigFile.SetActiveProfile(selected)
					v.StatusMessage = fmt.Sprintf("✓ Switched active profile to %q", selected)
					return true, true, false
				}

			case "n", "N":
				v.Mode = profileModeNew
				v.NameInput.SetValue("")
				v.NameInput.Focus()
				v.StatusMessage = ""
				return false, false, false

			case "x", "X", "d", "D", "delete":
				if len(names) <= 1 {
					v.StatusMessage = "⚠️ Cannot delete the only profile."
					return false, false, false
				}
				if v.Cursor >= 0 && v.Cursor < len(names) {
					v.DeleteTarget = names[v.Cursor]
					v.Mode = profileModeConfirmDelete
					v.StatusMessage = ""
				}
				return false, false, false

			case "esc", "q", "P", "p":
				return false, false, true
			}
		}
	}
	return false, false, false
}

// View renders the profile management dialog.
func (v *ProfilesView) View() string {
	contentWidth := style.ClampWidth(v.Width, 10, 65, 95)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render("📋 GAME PROFILES MANAGER")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Each profile stores a complete, independent execution configuration in rpt.toml")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	if v.StatusMessage != "" {
		b.WriteString(lipgloss.NewStyle().Foreground(style.ColorHighlight).Render(v.StatusMessage) + "\n\n")
	}

	names := v.ConfigFile.ProfileNames()

	switch v.Mode {
	case profileModeNew:
		b.WriteString(lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render("Create New Profile (Cloned from selected profile):") + "\n\n")
		b.WriteString("Profile Name: " + v.NameInput.View() + "\n\n")
		b.WriteString(style.KeyStyle.Render("[Enter] Save Profile  •  [Esc] Cancel") + "\n")

	case profileModeConfirmDelete:
		warnStyle := lipgloss.NewStyle().Bold(true).Foreground(style.ColorDanger)
		b.WriteString(warnStyle.Render(fmt.Sprintf("⚠️ Delete Profile %q?", v.DeleteTarget)) + "\n\n")
		b.WriteString("This action will remove all custom settings for this profile.\n\n")
		b.WriteString(style.KeyStyle.Render("[y/Enter] Confirm Delete  •  [n/Esc] Cancel") + "\n")

	case profileModeList:
		for i, name := range names {
			prof := v.ConfigFile.Profiles[name]
			cursor := "  "
			if i == v.Cursor {
				cursor = "❯ "
			}

			isActive := name == v.ConfigFile.ActiveProfile

			nameBadge := lipgloss.NewStyle().Bold(true).Foreground(style.ColorPrimary).Render(name)
			if i == v.Cursor {
				nameBadge = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render(name)
			}

			statusBadge := ""
			if isActive {
				statusBadge = " " + lipgloss.NewStyle().
					Bold(true).
					Foreground(lipgloss.Color("#11111b")).
					Background(style.ColorSuccess).
					Padding(0, 1).
					Render("ACTIVE")
			}

			// Profile Summary Info
			exeInfo := "None"
			if prof != nil && prof.TargetExe != "" {
				exeInfo = prof.TargetExe
			}
			gsInfo := "Gamescope: OFF"
			if prof != nil && prof.UseGamescope {
				if prof.GamescopeWidth > 0 && prof.GamescopeHeight > 0 {
					gsInfo = fmt.Sprintf("Gamescope: %dx%d", prof.GamescopeWidth, prof.GamescopeHeight)
				} else {
					gsInfo = "Gamescope: Auto"
				}
			}

			b.WriteString(fmt.Sprintf("%s%s%s\n", cursor, nameBadge, statusBadge))
			b.WriteString(fmt.Sprintf("    └─ Exe: %s • %s\n\n", exeInfo, gsInfo))
		}

		b.WriteString(divider + "\n")
		b.WriteString(style.KeyStyle.Render("[↑/↓] Navigate  •  [Enter] Activate  •  [n] New Profile  •  [x] Delete  •  [Esc] Back") + "\n")
	}

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorBorder).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
