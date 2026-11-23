package views

import (
	"cmp"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/broli/run-proton-tui/internal/pathutil"
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
	AllItems       []ExeItem
	FilteredItems  []ExeItem
	Cursor         int
	Selected       string
	HideInstallers bool
}

// DiscoverExecutables scans the game directory (up to 4 levels) for .exe files.
func DiscoverExecutables(gameDir string) []ExeItem {
	var items []ExeItem
	baseDepth := strings.Count(gameDir, string(os.PathSeparator))

	_ = filepath.WalkDir(gameDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if d.IsDir() {
			name := d.Name()
			if pathutil.IsIgnoredDir(name) {
				return filepath.SkipDir
			}
			if strings.Count(path, string(os.PathSeparator))-baseDepth > 4 {
				return filepath.SkipDir
			}
			return nil
		}

		if strings.HasSuffix(strings.ToLower(d.Name()), ".exe") {
			info, err := d.Info()
			if err != nil {
				return nil
			}
			rel, _ := filepath.Rel(gameDir, path)
			items = append(items, ExeItem{
				RelativePath:   rel,
				Size:           info.Size(),
				Classification: runner.ClassifyExecutable(rel),
			})
		}
		return nil
	})

	// Sort executables intelligently:
	// 1. 3D games come before 2D utilities
	// 2. Binaries in game/ or games/ directory come earlier
	// 3. Binaries matching directory name come earlier
	// 4. Larger file size comes earlier
	dirBase := strings.ToLower(filepath.Base(gameDir))
	slices.SortStableFunc(items, func(a, b ExeItem) int {
		aIs3D := a.Classification.Type == runner.ExeType3DGame
		bIs3D := b.Classification.Type == runner.ExeType3DGame
		if aIs3D != bIs3D {
			if aIs3D {
				return -1
			}
			return 1
		}

		aRel := strings.ToLower(a.RelativePath)
		bRel := strings.ToLower(b.RelativePath)
		aInGameDir := strings.HasPrefix(aRel, "games/") || strings.HasPrefix(aRel, "game/")
		bInGameDir := strings.HasPrefix(bRel, "games/") || strings.HasPrefix(bRel, "game/")
		if aInGameDir != bInGameDir {
			if aInGameDir {
				return -1
			}
			return 1
		}

		aBase := strings.ToLower(filepath.Base(a.RelativePath))
		bBase := strings.ToLower(filepath.Base(b.RelativePath))
		aMatch := strings.Contains(aBase, dirBase)
		bMatch := strings.Contains(bBase, dirBase)
		if aMatch != bMatch {
			if aMatch {
				return -1
			}
			return 1
		}

		return cmp.Compare(b.Size, a.Size)
	})

	return items
}

// NewExePickerView initializes the picker.
func NewExePickerView(items []ExeItem, currentExe string) *ExePickerView {
	picker := &ExePickerView{
		AllItems:       items,
		HideInstallers: false,
		Selected:       currentExe,
	}
	picker.recomputeFiltered()
	return picker
}

func (p *ExePickerView) recomputeFiltered() {
	if !p.HideInstallers {
		p.FilteredItems = p.AllItems
	} else {
		var filtered []ExeItem
		for _, it := range p.AllItems {
			base := strings.ToLower(filepath.Base(it.RelativePath))
			isSetup := strings.Contains(base, "setup") ||
				strings.Contains(base, "unins") ||
				strings.Contains(base, "installer") ||
				strings.Contains(base, "patch") ||
				strings.Contains(base, "crashreport")

			if !isSetup || it.RelativePath == p.Selected {
				filtered = append(filtered, it)
			}
		}
		p.FilteredItems = filtered
	}

	p.Cursor = 0
	for i, it := range p.FilteredItems {
		if it.RelativePath == p.Selected {
			p.Cursor = i
			break
		}
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
			if p.Cursor < len(p.FilteredItems)-1 {
				p.Cursor++
			}
		case "i", "I":
			p.HideInstallers = !p.HideInstallers
			p.recomputeFiltered()
			return p, false, false
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			idx := int(msg.String()[0] - '1')
			if idx < len(p.FilteredItems) {
				p.Cursor = idx
				p.Selected = p.FilteredItems[idx].RelativePath
				return p, true, false
			}
		case "enter", " ":
			if len(p.FilteredItems) > 0 {
				p.Selected = p.FilteredItems[p.Cursor].RelativePath
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
	sb.WriteString(style.TitleStyle.Render("Select Game Executable (Press [1-9] or [Enter]/[Space] to Select, [i] Filter Setups, [Esc] to Cancel)"))
	sb.WriteString("\n\n")

	if len(p.FilteredItems) == 0 {
		sb.WriteString(style.BadgeWarning.Render("No matching .exe binaries found in this directory."))
		return sb.String()
	}

	for i, item := range p.FilteredItems {
		cursor := "  "
		lineStyle := style.ItemStyle
		if i == p.Cursor {
			cursor = "👉"
			lineStyle = style.SelectedItemStyle
		}

		numKey := ""
		if i < 9 {
			numKey = style.KeyBadge.Render(fmt.Sprintf("[%d]", i+1)) + " "
		}

		tag := style.BadgeHighlight.Render("3D Game")
		if item.Classification.Type == runner.ExeType2DUtility {
			tag = style.BadgeWarning.Render("2D Utility / Setup")
		}

		sizeStr := fmt.Sprintf("%.1f MB", float64(item.Size)/(1024*1024))
		sb.WriteString(fmt.Sprintf("%s %s%s (%s) %s\n", cursor, numKey, lineStyle.Render(item.RelativePath), sizeStr, tag))
	}

	sb.WriteString("\n")
	filterStatus := "Show All (Installers Visible)"
	if p.HideInstallers {
		filterStatus = "Filtered (Installers Hidden)"
	}
	sb.WriteString(lipgloss.NewStyle().Foreground(style.ColorMuted).Render(fmt.Sprintf("[i] %s", filterStatus)))

	return sb.String()
}
