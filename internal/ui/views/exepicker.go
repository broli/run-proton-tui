package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/runner"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// ExeItem represents a selectable Windows executable.
type ExeItem struct {
	RelativePath   string
	Size           int64
	Classification runner.ClassificationResult
}

// ExePickerView handles interactive executable selection.
type ExePickerView struct {
	Items    []ExeItem
	Cursor   int
	Selected string
}

// DiscoverExecutables scans the game directory (up to 4 levels) for .exe files.
func DiscoverExecutables(gameDir string) []ExeItem {
	var items []ExeItem
	baseDepth := strings.Count(gameDir, string(os.PathSeparator))

	_ = filepath.Walk(gameDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			name := info.Name()
			if name == "proton-prefix" || name == ".logs" || name == ".git" {
				return filepath.SkipDir
			}
			if strings.Count(path, string(os.PathSeparator))-baseDepth > 4 {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(strings.ToLower(info.Name()), ".exe") {
			rel, _ := filepath.Rel(gameDir, path)
			items = append(items, ExeItem{
				RelativePath:   rel,
				Size:           info.Size(),
				Classification: runner.ClassifyExecutable(rel),
			})
		}
		return nil
	})

	return items
}

// NewExePickerView initializes the picker.
func NewExePickerView(items []ExeItem, currentExe string) *ExePickerView {
	cursor := 0
	for i, it := range items {
		if it.RelativePath == currentExe {
			cursor = i
			break
		}
	}
	return &ExePickerView{
		Items:    items,
		Cursor:   cursor,
		Selected: currentExe,
	}
}

// Update handles navigation keys in the executable picker.
func (p *ExePickerView) Update(msg tea.Msg) (*ExePickerView, bool, bool) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "up", "k":
			if p.Cursor > 0 {
				p.Cursor--
			}
		case "down", "j":
			if p.Cursor < len(p.Items)-1 {
				p.Cursor++
			}
		case "enter":
			if len(p.Items) > 0 {
				p.Selected = p.Items[p.Cursor].RelativePath
				return p, true, false // selected, don't cancel
			}
		case "esc", "q":
			return p, false, true // cancel
		}
	}
	return p, false, false
}

// View renders the executable list.
func (p *ExePickerView) View() string {
	var sb strings.Builder
	sb.WriteString(style.TitleStyle.Render("🎯 Select Game Executable (Press [Enter] to Select, [Esc] to Cancel)"))
	sb.WriteString("\n\n")

	if len(p.Items) == 0 {
		sb.WriteString(style.BadgeWarning.Render("No .exe binaries found in this directory."))
		return sb.String()
	}

	for i, item := range p.Items {
		cursor := "  "
		lineStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == p.Cursor {
			cursor = "👉"
			lineStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		tag := style.BadgeHighlight.Render("3D Game")
		if item.Classification.Type == runner.ExeType2DUtility {
			tag = style.BadgeWarning.Render("2D Utility")
		}

		sizeStr := fmt.Sprintf("%.1f MB", float64(item.Size)/(1024*1024))
		sb.WriteString(fmt.Sprintf("%s %s (%s) %s\n", cursor, lineStyle.Render(item.RelativePath), sizeStr, tag))
	}

	return sb.String()
}
