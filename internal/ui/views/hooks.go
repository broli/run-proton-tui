package views

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/hooks"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HooksView handles lifecycle hook inspection, cascade resolution, and syntax-highlighted paging.
type HooksView struct {
	GameDir          string
	Config           *config.GameConfig
	Report           hooks.FullHooksReport
	ViewingScript    bool
	ActiveScriptPath string
	Viewport         viewport.Model
	Width            int
	Height           int
	StatusMessage    string
}

// NewHooksView initializes the hooks inspector view.
func NewHooksView(gameDir string, cfg *config.GameConfig, width, height int) *HooksView {
	opts := hooks.HookOptions{
		GameDir:        gameDir,
		PrefixDir:      filepath.Join(gameDir, ".prefix"),
		TargetExe:      cfg.TargetExe,
		ProtonPath:     cfg.ProtonPath,
		ConfigHookPath: cfg.PreLaunchHook,
		ExtraHookDirs:  cfg.HookDirs,
	}
	if cfg.UseGamescope {
		opts.GamescopeDisplay = ":1"
	}
	report := hooks.InspectAllHooks(opts)

	vp := viewport.New(width-4, height-6)
	return &HooksView{
		GameDir:       gameDir,
		Config:        cfg,
		Report:        report,
		ViewingScript: false,
		Viewport:      vp,
		Width:         width,
		Height:        height,
	}
}

// SetDimensions updates the viewport dimensions on window resize.
func (h *HooksView) SetDimensions(width, height int) {
	h.Width = width
	h.Height = height
	h.Viewport.Width = width - 4
	h.Viewport.Height = height - 6
}

// Update handles key navigation within the hooks inspector and script pager.
func (h *HooksView) Update(msg tea.Msg) (*HooksView, bool) {
	if h.ViewingScript {
		switch msg := msg.(type) {
		case tea.KeyMsg:
			switch msg.String() {
			case "esc", "q", "h", "H", "backspace":
				h.ViewingScript = false
				h.StatusMessage = ""
				return h, false
			}
		}
		var cmd tea.Cmd
		h.Viewport, cmd = h.Viewport.Update(msg)
		_ = cmd
		return h, false
	}

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "1", "p", "P":
			return h.openScript(h.Report.PreLaunch.Resolved, "pre_launch.sh")
		case "2", "e", "E":
			return h.openScript(h.Report.PostExit.Resolved, "post_exit.sh")
		case "esc", "q", "backspace":
			return h, true // Return to previous view / dashboard
		}
	}
	return h, false
}

func (h *HooksView) openScript(path, hookName string) (*HooksView, bool) {
	if path == "" {
		h.StatusMessage = fmt.Sprintf("No %s script found in search paths. Create one in ./hooks/%s or ask an AI helper!", hookName, hookName)
		return h, false
	}

	data, err := os.ReadFile(path)
	if err != nil {
		h.StatusMessage = fmt.Sprintf("Failed to read script %s: %v", filepath.Base(path), err)
		return h, false
	}

	highlighted, err := hooks.HighlightBash(string(data))
	if err != nil {
		highlighted = string(data)
	}

	h.ActiveScriptPath = path
	h.Viewport.SetContent(highlighted)
	h.Viewport.GotoTop()
	h.ViewingScript = true
	h.StatusMessage = ""
	return h, false
}

// View renders either the syntax-highlighted script pager or the hook inspector overview.
func (h *HooksView) View() string {
	if h.ViewingScript {
		header := style.TitleStyle.Render(fmt.Sprintf("📜 Viewing Hook: %s (Syntax Highlighted with Chroma)", h.ActiveScriptPath))
		footer := style.SubheaderStyle.Render(fmt.Sprintf("Scroll: ↑/k, ↓/j, PgUp, PgDown • Progress: %3.f%% • [Esc/q] Return to Inspector", h.Viewport.ScrollPercent()*100))
		return lipgloss.JoinVertical(lipgloss.Left, header, h.Viewport.View(), footer)
	}

	contentWidth := h.Width - 6
	if contentWidth < 70 {
		contentWidth = 70
	}
	if contentWidth > 110 {
		contentWidth = 110
	}

	var sb strings.Builder

	// Header
	header := style.TitleStyle.Render("🔗 LIFECYCLE HOOK INSPECTOR & PRESENCE REPORT")
	desc := style.SubheaderStyle.Render("Cascading resolution for game-specific workarounds (patches, memory tweaks, save sync)")
	divider := lipgloss.NewStyle().Foreground(style.ColorBorder).Render(strings.Repeat("─", contentWidth))

	sb.WriteString(header + "\n")
	sb.WriteString(desc + "\n")
	sb.WriteString(divider + "\n\n")

	if h.StatusMessage != "" {
		sb.WriteString(style.BadgeWarning.Render(" ℹ "+h.StatusMessage) + "\n\n")
	}

	// 1. Hook Status Cards
	preBadge := style.BadgeMuted.Render("Not Found (Standard Launch)")
	if h.Report.PreLaunch.Resolved != "" {
		preBadge = style.BadgeSuccess.Render("ACTIVE: " + h.Report.PreLaunch.Resolved)
	}
	postBadge := style.BadgeMuted.Render("Not Found (No Post Cleanup)")
	if h.Report.PostExit.Resolved != "" {
		postBadge = style.BadgeSuccess.Render("ACTIVE: " + h.Report.PostExit.Resolved)
	}

	sb.WriteString(style.HeaderStyle.Render("📌 Lifecycle Phase Resolution:") + "\n")
	sb.WriteString(fmt.Sprintf("  %s %-24s %s\n", style.KeyBadge.Render("[1 / p]"), "Pre-Launch Hook:", preBadge))
	sb.WriteString(fmt.Sprintf("  %s %-24s %s\n\n", style.KeyBadge.Render("[2 / e]"), "Post-Exit Hook:", postBadge))

	// 2. Cascade Hierarchy Checklist
	sb.WriteString(style.HeaderStyle.Render("🔍 Cascade Search Checklist (Evaluated in Order):") + "\n")
	renderChecklist := func(candidates []hooks.CascadeCandidate) {
		for _, c := range candidates {
			marker := style.BadgeMuted.Render("[ ] Not Found")
			if c.Exists {
				if c.Selected {
					marker = style.BadgeSuccess.Render("[✓] Active")
				} else {
					marker = style.BadgeWarning.Render("[•] Shadowed")
				}
			}
			relPath := c.Path
			if rel, err := filepath.Rel(h.GameDir, c.Path); err == nil && !strings.HasPrefix(rel, "..") {
				relPath = "./" + rel
			}
			sb.WriteString(fmt.Sprintf("    %-14s %-32s (%s)\n", marker, relPath, c.Source))
		}
	}
	sb.WriteString(style.SubheaderStyle.Render("  Pre-Launch Search Order:") + "\n")
	renderChecklist(h.Report.PreLaunch.Candidates)
	sb.WriteString("\n")

	// 3. Exported Environment Variables Cheatsheet
	sb.WriteString(style.HeaderStyle.Render("📦 Injected Environment Variables ($RPT_* Cheatsheet):") + "\n")
	for _, env := range h.Report.EnvVars {
		nameStr := lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight).Render(fmt.Sprintf("$%-22s", env.Name))
		valStr := lipgloss.NewStyle().Foreground(style.ColorMuted).Render(fmt.Sprintf("= %s", env.CurrentVal))
		sb.WriteString(fmt.Sprintf("    %s %s\n", nameStr, valStr))
		sb.WriteString(fmt.Sprintf("      └─ %s\n", lipgloss.NewStyle().Foreground(style.ColorText).Render(env.Description)))
	}
	sb.WriteString("\n")

	// Dock / Navigation hints
	sb.WriteString(divider + "\n")
	dock := fmt.Sprintf("%s    %s",
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[1/2] View Script in Syntax-Highlighted Pager"),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[Esc / q] Return to Dashboard"),
	)
	sb.WriteString(dock)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorPrimary).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(sb.String())
}
