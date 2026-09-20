package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/prefix"
	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/runner"
	"github.com/broli/run-proton-tui/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/x/term"
)

// ViewState represents the currently active sub-screen in the TUI.
type ViewState int

const (
	StateDashboard ViewState = iota
	StateHelp
	StateExePicker
	StateProtonPicker
	StateOverrides
	StateDiagnostics
	StateLogs
)

// Model is the root Elm Architecture model for rpt.
type Model struct {
	State           ViewState
	GameDir         string
	GameTitle       string
	Config          *config.GameConfig
	Runners         []proton.Runner
	Exes            []views.ExeItem
	Emulator        *integrations.EmulatorInfo
	ProtonDB        *integrations.ProtonDBReport
	ActiveOverrides map[string]string
	GPUInfo         *hardware.GPUInfo
	CPUTopo         *hardware.CPUTopology
	StatusMessage   string
	Width           int
	Height          int
	ShouldLaunch    bool

	// Sub-views
	HelpView        *views.HelpView
	ExePickerView   *views.ExePickerView
	ProtonPickerView *views.ProtonPickerView
	OverridesView   *views.OverridesView
	DiagnosticsView *views.DiagnosticsView
	LogsView        *views.LogsView
}

type protonDBMsg *integrations.ProtonDBReport

// NewModel constructs the root application state.
func NewModel(gameDir string, cfg *config.GameConfig) (*Model, error) {
	runners, _ := proton.DiscoverRunners()
	exes := views.DiscoverExecutables(gameDir)
	emuInfo, _ := integrations.ScanEmulators(gameDir)
	gpuInfo, _ := hardware.DetectGPU()
	cpuTopo, _ := hardware.DetectCPUTopology()

	// If target_exe is empty but exes were discovered, pick the first
	if cfg.TargetExe == "" && len(exes) > 0 {
		cfg.TargetExe = exes[0].RelativePath
	}

	// If proton_path is empty but runners exist, pick the first
	if cfg.ProtonPath == "" && len(runners) > 0 {
		cfg.ProtonPath = runners[0].Path
	}

	// Read active overrides
	pfxDir := filepath.Join(gameDir, "proton-prefix")
	activeOverrides, _ := proton.ReadRegistryOverrides(pfxDir)

	termWidth, termHeight, err := term.GetSize(os.Stdout.Fd())
	if err != nil || termWidth < 80 {
		termWidth = 110
		termHeight = 32
	}

	m := &Model{
		State:           StateDashboard,
		GameDir:         gameDir,
		GameTitle:       filepath.Base(gameDir),
		Config:          cfg,
		Runners:         runners,
		Exes:            exes,
		Emulator:        emuInfo,
		ProtonDB:        nil,
		ActiveOverrides: activeOverrides,
		GPUInfo:         gpuInfo,
		CPUTopo:         cpuTopo,
		Width:           termWidth,
		Height:          termHeight,
		ShouldLaunch:    false,
	}

	return m, nil
}

// Init starts initial async background tasks (e.g. ProtonDB lookup).
func (m *Model) Init() tea.Cmd {
	appID := m.Config.AppID
	if (appID == "" || appID == "0") && m.Emulator != nil && m.Emulator.AppID != "" && m.Emulator.AppID != "0" {
		appID = m.Emulator.AppID
		m.Config.AppID = appID
	}

	if appID != "" && appID != "0" {
		return func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			if rep, err := integrations.FetchProtonDBReport(ctx, appID); err == nil {
				return protonDBMsg(rep)
			}
			return nil
		}
	}
	return nil
}

// Update processes incoming UI events and keybindings.
func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.Width = msg.Width
		m.Height = msg.Height
		if m.HelpView != nil {
			m.HelpView.Viewport.Width = msg.Width
			m.HelpView.Viewport.Height = msg.Height - 6
		}
		return m, nil

	case protonDBMsg:
		if msg != nil {
			m.ProtonDB = (*integrations.ProtonDBReport)(msg)
		}
		return m, nil

	case tea.KeyMsg:
		// Sub-view specific handling
		switch m.State {
		case StateHelp:
			if msg.String() == "?" || msg.String() == "esc" || msg.String() == "q" {
				m.State = StateDashboard
				return m, nil
			}
			var cmd tea.Cmd
			m.HelpView, cmd = m.HelpView.Update(msg)
			return m, cmd

		case StateExePicker:
			var selected, cancel bool
			m.ExePickerView, selected, cancel = m.ExePickerView.Update(msg)
			if selected {
				m.Config.TargetExe = m.ExePickerView.Selected
				// Auto-apply classification recommendations
				class := runner.ClassifyExecutable(m.Config.TargetExe)
				m.Config.UseGamescope = class.RecommendGamescope
				m.Config.UsePrimeRun = class.RecommendPrimeRun
				m.Config.UsePCores = class.RecommendPCores
				_ = config.SaveConfig(m.GameDir, m.Config)
				m.State = StateDashboard
			} else if cancel {
				m.State = StateDashboard
			}
			return m, nil

		case StateProtonPicker:
			var selected, cancel bool
			m.ProtonPickerView, selected, cancel = m.ProtonPickerView.Update(msg)
			if selected {
				m.Config.ProtonPath = m.ProtonPickerView.Selected.Path
				_ = config.SaveConfig(m.GameDir, m.Config)
				m.State = StateDashboard
			} else if cancel {
				m.State = StateDashboard
			}
			return m, nil

		case StateOverrides:
			var done bool
			m.OverridesView, done = m.OverridesView.Update(msg)
			if done {
				pfxDir := filepath.Join(m.GameDir, "proton-prefix")
				m.ActiveOverrides, _ = proton.ReadRegistryOverrides(pfxDir)
				m.State = StateDashboard
			}
			return m, nil

		case StateDiagnostics:
			var done bool
			m.DiagnosticsView, done = m.DiagnosticsView.Update(msg)
			if done {
				m.State = StateDashboard
			}
			return m, nil

		case StateLogs:
			var done bool
			m.LogsView, done = m.LogsView.Update(msg)
			if done {
				m.State = StateDashboard
			}
			return m, nil

		case StateDashboard:
			return m.handleDashboardKeys(msg)
		}
	}

	return m, nil
}

func (m *Model) handleDashboardKeys(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter", "1":
		m.ShouldLaunch = true
		_ = config.SaveConfig(m.GameDir, m.Config)
		return m, tea.Quit

	case "2":
		m.ProtonPickerView = views.NewProtonPickerView(m.Runners, m.Config.ProtonPath)
		m.State = StateProtonPicker
		return m, nil

	case "3":
		m.ExePickerView = views.NewExePickerView(m.Exes, m.Config.TargetExe)
		m.State = StateExePicker
		return m, nil

	case "g":
		m.Config.UseGamescope = !m.Config.UseGamescope
		_ = config.SaveConfig(m.GameDir, m.Config)
		return m, nil

	case "p":
		m.Config.UsePCores = !m.Config.UsePCores
		_ = config.SaveConfig(m.GameDir, m.Config)
		return m, nil

	case "v":
		m.Config.UsePrimeRun = !m.Config.UsePrimeRun
		_ = config.SaveConfig(m.GameDir, m.Config)
		return m, nil

	case "o":
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		m.OverridesView = views.NewOverridesView(pfxDir, m.Config.ProtonPath, m.ActiveOverrides)
		m.State = StateOverrides
		return m, nil

	case "a":
		// Query ProtonDB on-demand
		if m.Config.AppID != "" && m.Config.AppID != "0" {
			m.StatusMessage = "Querying ProtonDB API..."
			return m, func() tea.Msg {
				ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
				defer cancel()
				if rep, err := integrations.FetchProtonDBReport(ctx, m.Config.AppID); err == nil {
					return protonDBMsg(rep)
				}
				return nil
			}
		}
		// If AppID is 0, try Steam store search
		m.StatusMessage = "Searching Steam for AppID..."
		return m, func() tea.Msg {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Second)
			defer cancel()
			if aid, _, err := integrations.SearchSteamAppID(ctx, m.GameTitle); err == nil {
				m.Config.AppID = aid
				_ = config.SaveConfig(m.GameDir, m.Config)
				if rep, err := integrations.FetchProtonDBReport(ctx, aid); err == nil {
					return protonDBMsg(rep)
				}
			}
			return nil
		}

	case "c":
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		res, err := prefix.SafeCleanPrefix(pfxDir, m.Config.ProtonPath, m.GameDir, m.Config.TargetExe, m.GameTitle)
		if err != nil {
			m.StatusMessage = fmt.Sprintf("Clean failed: %v", err)
		} else if res.BackupDir != "" {
			m.StatusMessage = fmt.Sprintf("Prefix cleaned! Saves backed up to: %s", res.BackupDir)
		} else {
			m.StatusMessage = "Prefix cleaned and reset successfully."
		}
		return m, nil

	case "h":
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		report := diagnostics.RunPreflightCheck(m.GameDir, m.Config.TargetExe, pfxDir)
		m.DiagnosticsView = views.NewDiagnosticsView(report)
		m.State = StateDiagnostics
		return m, nil

	case "l":
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		logs := diagnostics.FindRecentLogs(m.GameDir, pfxDir)
		m.LogsView = views.NewLogsView(logs, m.Width, m.Height)
		m.State = StateLogs
		return m, nil

	case "?", "f1":
		m.HelpView = views.NewHelpView(m.Width, m.Height)
		m.State = StateHelp
		return m, nil

	case "b":
		m.Config.OpaqueBackdrop = !m.Config.OpaqueBackdrop
		_ = config.SaveConfig(m.GameDir, m.Config)
		return m, nil

	case "q", "ctrl+c":
		return m, tea.Quit
	}

	return m, nil
}

// View renders the currently active sub-screen.
func (m *Model) View() string {
	switch m.State {
	case StateHelp:
		if m.HelpView != nil {
			return m.HelpView.View()
		}
	case StateExePicker:
		if m.ExePickerView != nil {
			return m.ExePickerView.View()
		}
	case StateProtonPicker:
		if m.ProtonPickerView != nil {
			return m.ProtonPickerView.View()
		}
	case StateOverrides:
		if m.OverridesView != nil {
			return m.OverridesView.View()
		}
	case StateDiagnostics:
		if m.DiagnosticsView != nil {
			return m.DiagnosticsView.View()
		}
	case StateLogs:
		if m.LogsView != nil {
			return m.LogsView.View()
		}
	case StateDashboard:
		pName := filepath.Base(filepath.Dir(m.Config.ProtonPath))
		if pName == "." || pName == "" {
			pName = filepath.Base(m.Config.ProtonPath)
		}
		return views.RenderDashboard(views.DashboardData{
			GameTitle:       m.GameTitle,
			GameDir:         m.GameDir,
			Config:          m.Config,
			Emulator:        m.Emulator,
			ProtonDB:        m.ProtonDB,
			Classification:  runner.ClassifyExecutable(m.Config.TargetExe),
			HasNTSync:       hardware.HasNTSync(),
			HasPrimeRun:     m.GPUInfo != nil && m.GPUInfo.HasPrimeRun,
			HasGamescope:    true,
			ActiveOverrides: m.ActiveOverrides,
			ProtonName:      pName,
			StatusMessage:   m.StatusMessage,
			Width:           m.Width,
			Height:          m.Height,
			OpaqueBackdrop:  m.Config.OpaqueBackdrop,
		})
	}
	return ""
}
