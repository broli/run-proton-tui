package ui

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/prefix"
	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/quirks"
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
	StatePresetPicker
	StateProtonDB
	StateSubMenu
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
	PrefixCleaned   bool
	Width           int
	Height          int
	ShouldLaunch    bool

	// Sub-views
	HelpView         *views.HelpView
	ExePickerView    *views.ExePickerView
	ProtonPickerView *views.ProtonPickerView
	OverridesView    *views.OverridesView
	DiagnosticsView  *views.DiagnosticsView
	LogsView         *views.LogsView
	PresetPickerView *views.PresetPickerView
	ProtonDBView     *views.ProtonDBView
	SubMenuView      *views.SubMenuView
	DetectedPreset   *quirks.Preset
}

type protonDBResultMsg struct {
	Report *integrations.ProtonDBReport
	AppID  string
	Err    error
}

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

	// Auto-detect known non-standard quirks preset on initial run if not already set
	if cfg.PresetName == "" {
		if p := quirks.DetectQuirks(gameDir, cfg.TargetExe, cfg.AppID, cfg.ProtonPath); p != nil {
			cfg.PresetName = p.Name
			if p.UmuID != "" && cfg.UmuID == "" {
				cfg.UmuID = p.UmuID
			}
			if p.DisplayFile != "" && cfg.DisplayFile == "" {
				cfg.DisplayFile = p.DisplayFile
			}
			if len(p.ExtraArgs) > 0 && len(cfg.ExtraArgs) == 0 {
				cfg.ExtraArgs = p.ExtraArgs
			}
			if len(p.WaitProcesses) > 0 && len(cfg.WaitProcesses) == 0 {
				cfg.WaitProcesses = p.WaitProcesses
			}
			for k, v := range p.EnvVars {
				if cfg.EnvVars == nil {
					cfg.EnvVars = make(map[string]string)
				}
				if _, ok := cfg.EnvVars[k]; !ok {
					cfg.EnvVars[k] = v
				}
			}
			for k, v := range p.Profiles {
				if cfg.Profiles == nil {
					cfg.Profiles = make(map[string]*config.ExecutableProfile)
				}
				if _, ok := cfg.Profiles[k]; !ok {
					cfg.Profiles[k] = v
				}
			}
		}
	}

	// Read active overrides
	pfxDir := filepath.Join(gameDir, "proton-prefix")
	activeOverrides, _ := proton.ReadRegistryOverrides(pfxDir)
	if activeOverrides == nil {
		activeOverrides = make(map[string]string)
	}
	for k, v := range cfg.DLLOverrides {
		if _, ok := activeOverrides[k]; !ok {
			activeOverrides[k] = v
		}
	}

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

func (m *Model) fetchProtonDB(customQuery string) tea.Cmd {
	m.StatusMessage = "Querying ProtonDB & Steam..."
	if m.ProtonDBView != nil {
		m.ProtonDBView.Status = "Querying ProtonDB & Steam API..."
	}

	appID := m.Config.AppID
	presetName := m.Config.PresetName
	targetExe := m.Config.TargetExe
	gameTitle := m.GameTitle

	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if customQuery != "" {
			if _, err := strconv.Atoi(customQuery); err == nil {
				appID = customQuery
			} else {
				aid, _, err := integrations.SearchSteamAppID(ctx, customQuery)
				if err != nil {
					return protonDBResultMsg{
						Err: fmt.Errorf("Steam search for %q: %w", customQuery, err),
					}
				}
				appID = aid
			}
		}

		if appID == "" || appID == "0" {
			aid, _, err := integrations.ResolveSteamAppID(ctx, presetName, targetExe, gameTitle)
			if err != nil {
				return protonDBResultMsg{
					Err: fmt.Errorf("no Steam title matched: %w", err),
				}
			}
			appID = aid
		}

		rep, err := integrations.FetchProtonDBReport(ctx, appID)
		if err != nil {
			return protonDBResultMsg{
				AppID: appID,
				Err:   fmt.Errorf("ProtonDB API (AppID %s): %w", appID, err),
			}
		}

		return protonDBResultMsg{
			Report: rep,
			AppID:  appID,
			Err:    nil,
		}
	}
}

// Init starts initial async background tasks (e.g. ProtonDB lookup).
func (m *Model) Init() tea.Cmd {
	if (m.Config.AppID == "" || m.Config.AppID == "0") && m.Emulator != nil && m.Emulator.AppID != "" && m.Emulator.AppID != "0" {
		m.Config.AppID = m.Emulator.AppID
	}

	return m.fetchProtonDB("")
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

	case protonDBResultMsg:
		if msg.Err != nil {
			m.StatusMessage = fmt.Sprintf("ProtonDB: %v", msg.Err)
			if m.ProtonDBView != nil {
				m.ProtonDBView.Status = fmt.Sprintf("%v", msg.Err)
			}
		} else if msg.Report != nil {
			m.ProtonDB = msg.Report
			m.Config.AppID = msg.AppID
			_ = config.SaveConfig(m.GameDir, m.Config)
			m.StatusMessage = fmt.Sprintf("ProtonDB: %s (%d reports)", msg.Report.GetTierBadge(), msg.Report.Total)
			if m.ProtonDBView != nil {
				m.ProtonDBView.Report = m.ProtonDB
				m.ProtonDBView.AppID = msg.AppID
				m.ProtonDBView.Status = ""
			}
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
				// Auto-apply per-executable profile or classification recommendations
				if _, ok := m.Config.Profiles[m.Config.TargetExe]; ok {
					eff := m.Config.GetEffectiveConfig(m.Config.TargetExe)
					m.Config.UseGamescope = eff.UseGamescope
					m.Config.UsePrimeRun = eff.UsePrimeRun
					m.Config.UsePCores = eff.UsePCores
					m.StatusMessage = fmt.Sprintf("Switched to %s profile!", m.Config.TargetExe)
				} else {
					class := runner.ClassifyExecutable(m.Config.TargetExe)
					m.Config.UseGamescope = class.RecommendGamescope
					m.Config.UsePrimeRun = class.RecommendPrimeRun
					m.Config.UsePCores = class.RecommendPCores
				}
				_ = config.SaveConfig(m.GameDir, m.Config)
				m.State = StateDashboard
			} else if cancel {
				m.State = StateDashboard
			}
			return m, nil

		case StatePresetPicker:
			if m.PresetPickerView != nil {
				applied, cancel := m.PresetPickerView.Update(msg)
				if applied {
					preset := m.PresetPickerView.Preset
					if preset != nil {
						m.Config.PresetName = preset.Name
						if preset.UmuID != "" {
							m.Config.UmuID = preset.UmuID
						}
						if preset.DisplayFile != "" {
							m.Config.DisplayFile = preset.DisplayFile
						}
						if len(preset.ExtraArgs) > 0 {
							m.Config.ExtraArgs = preset.ExtraArgs
						}
						if len(preset.WaitProcesses) > 0 {
							m.Config.WaitProcesses = preset.WaitProcesses
						}
						if m.Config.EnvVars == nil {
							m.Config.EnvVars = make(map[string]string)
						}
						for k, v := range preset.EnvVars {
							m.Config.EnvVars[k] = v
						}
						if m.Config.Profiles == nil {
							m.Config.Profiles = make(map[string]*config.ExecutableProfile)
						}
						for k, v := range preset.Profiles {
							m.Config.Profiles[k] = v
						}
						_ = config.SaveConfig(m.GameDir, m.Config)
						m.StatusMessage = fmt.Sprintf("Applied %s preset!", preset.Name)
					}
					m.State = StateDashboard
				} else if cancel {
					m.State = StateDashboard
				}
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
				m.Config.DLLOverrides = m.OverridesView.GetActiveOverrides()
				// Ensure active overrides reflect what was set
				for k, v := range m.Config.DLLOverrides {
					m.ActiveOverrides[k] = v
				}
				_ = config.SaveConfig(m.GameDir, m.Config)
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

		case StateProtonDB:
			if m.ProtonDBView != nil {
				customQuery, refresh, done, cmd := m.ProtonDBView.Update(msg)
				if done {
					m.State = StateDashboard
					return m, nil
				}
				if customQuery != "" {
					return m, m.fetchProtonDB(customQuery)
				}
				if refresh {
					return m, m.fetchProtonDB("")
				}
				return m, cmd
			}
			return m, nil

		case StateSubMenu:
			if m.SubMenuView != nil {
				act := m.SubMenuView.Update(msg)
				return m.handleSubMenuAction(act)
			}
			return m, nil

		case StateDashboard:
			return m.handleDashboardKeys(msg)
		}
	}

	return m, nil
}

func (m *Model) openSubMenu(menuType views.SubMenuType) (tea.Model, tea.Cmd) {
	m.SubMenuView = views.NewSubMenuView(menuType, m.getSubMenuData())
	m.State = StateSubMenu
	return m, nil
}

func (m *Model) getSubMenuData() views.SubMenuData {
	pName := filepath.Base(filepath.Dir(m.Config.ProtonPath))
	if pName == "." || pName == "" {
		pName = filepath.Base(m.Config.ProtonPath)
	}

	pfxDriveC := filepath.Join(m.GameDir, "proton-prefix", "pfx", "drive_c")
	_, err := os.Stat(pfxDriveC)
	pfxActive := err == nil

	return views.SubMenuData{
		Config:          m.Config,
		ProtonName:      pName,
		ProtonDB:        m.ProtonDB,
		HasPrimeRun:     m.GPUInfo != nil && m.GPUInfo.HasPrimeRun,
		HasGamescope:    true,
		ActiveOverrides: m.ActiveOverrides,
		OpaqueBackdrop:  m.Config.OpaqueBackdrop,
		StatusMessage:   m.StatusMessage,
		PrefixActive:    pfxActive,
		PrefixCleaned:   m.PrefixCleaned,
		Width:           m.Width,
		Height:          m.Height,
	}
}

func (m *Model) updateSubMenuData() {
	if m.SubMenuView != nil {
		m.SubMenuView.Data = m.getSubMenuData()
	}
}

func (m *Model) handleSubMenuAction(act views.SubMenuAction) (tea.Model, tea.Cmd) {
	switch act {
	case views.ActionClose:
		m.State = StateDashboard
		return m, nil

	case views.ActionOpenProtonPicker:
		m.ProtonPickerView = views.NewProtonPickerView(m.Runners, m.Config.ProtonPath)
		m.State = StateProtonPicker
		return m, nil

	case views.ActionOpenExePicker:
		m.ExePickerView = views.NewExePickerView(m.Exes, m.Config.TargetExe)
		m.State = StateExePicker
		return m, nil

	case views.ActionOpenQuirks:
		p := quirks.GetActiveOrDetectedPreset(m.GameDir, m.Config.TargetExe, m.Config.AppID, m.Config.ProtonPath, m.Config)
		m.DetectedPreset = p
		m.PresetPickerView = views.NewPresetPickerView(p, m.GameTitle, m.Config, m.Width, m.Height)
		m.State = StatePresetPicker
		return m, nil

	case views.ActionOpenProtonDB:
		m.ProtonDBView = views.NewProtonDBView(m.ProtonDB, m.GameTitle, m.Config.AppID, m.Width, m.Height)
		m.State = StateProtonDB
		if m.ProtonDB == nil {
			return m, m.fetchProtonDB("")
		}
		return m, nil

	case views.ActionToggleGamescope:
		m.Config.UseGamescope = !m.Config.UseGamescope
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.UseGamescope {
			m.StatusMessage = fmt.Sprintf("Gamescope ENABLED (%dx%d @ %dHz -> %s)", m.Config.GamescopeWidth, m.Config.GamescopeHeight, m.Config.GamescopeRefresh, m.Config.GamescopeOutput)
		} else {
			m.StatusMessage = "Gamescope DISABLED (Native Window)"
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionTogglePCores:
		m.Config.UsePCores = !m.Config.UsePCores
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.UsePCores {
			m.StatusMessage = fmt.Sprintf("CPU P-Core Pinning ENABLED (Threads %s)", m.Config.PCoresMask)
		} else {
			m.StatusMessage = "CPU P-Core Pinning DISABLED (All Threads)"
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionTogglePrimeRun:
		m.Config.UsePrimeRun = !m.Config.UsePrimeRun
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.UsePrimeRun {
			m.StatusMessage = "GPU Runner: prime-run (NVIDIA RTX Dedicated)"
		} else {
			m.StatusMessage = "GPU Runner: Host iGPU (Power Saving)"
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionToggleXalia:
		m.Config.UseXalia = !m.Config.UseXalia
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.UseXalia {
			m.StatusMessage = "Proton Xalia UI Bridge ENABLED"
		} else {
			m.StatusMessage = "Proton Xalia UI Bridge DISABLED"
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionCycleDisplayOutput:
		outputs := hardware.GetConnectedDisplayOutputs()
		options := append([]string{"auto"}, outputs...)
		curIdx := 0
		for i, o := range options {
			if strings.EqualFold(o, m.Config.GamescopeOutput) {
				curIdx = i
				break
			}
		}
		nextIdx := (curIdx + 1) % len(options)
		m.Config.GamescopeOutput = options[nextIdx]
		_ = config.SaveConfig(m.GameDir, m.Config)
		m.StatusMessage = fmt.Sprintf("Display output set to: %s", m.Config.GamescopeOutput)
		m.updateSubMenuData()
		return m, nil

	case views.ActionOpenOverrides:
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		m.OverridesView = views.NewOverridesView(pfxDir, m.Config.ProtonPath, m.ActiveOverrides, m.Config.DLLOverrides)
		m.State = StateOverrides
		return m, nil

	case views.ActionCleanPrefix:
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		res, err := prefix.SafeCleanPrefixCustom(pfxDir, m.Config.ProtonPath, m.GameDir, m.Config.TargetExe, m.GameTitle, m.Config.BackupDir, m.Config.Filesystem.ExtraBackupPaths)
		if err != nil {
			m.StatusMessage = fmt.Sprintf("Clean failed: %v", err)
			m.PrefixCleaned = false
		} else if res.BackupDir != "" {
			m.StatusMessage = fmt.Sprintf("Prefix cleaned! Saved data preserved to: %s", res.BackupDir)
			m.PrefixCleaned = true
		} else {
			m.StatusMessage = "Prefix cleaned and reset successfully."
			m.PrefixCleaned = true
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionOpenDiagnostics:
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		report := diagnostics.RunPreflightCheck(m.GameDir, m.Config.TargetExe, pfxDir)
		m.DiagnosticsView = views.NewDiagnosticsView(report)
		m.State = StateDiagnostics
		return m, nil

	case views.ActionToggleLogging:
		m.Config.EnableLogging = !m.Config.EnableLogging
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.EnableLogging {
			m.StatusMessage = "Session logging ENABLED (writes to .logs/ on launch)"
		} else {
			m.StatusMessage = "Session logging DISABLED"
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionOpenLogs:
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		logs := diagnostics.FindRecentLogs(m.GameDir, pfxDir)
		m.LogsView = views.NewLogsView(logs, m.Width, m.Height)
		m.State = StateLogs
		return m, nil

	case views.ActionToggleBackdrop:
		m.Config.OpaqueBackdrop = !m.Config.OpaqueBackdrop
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.OpaqueBackdrop {
			m.StatusMessage = "Backdrop: Solid Dark"
		} else {
			m.StatusMessage = "Backdrop: Transparent"
		}
		m.updateSubMenuData()
		return m, nil

	case views.ActionOpenHelp:
		m.HelpView = views.NewHelpView(m.Width, m.Height)
		m.State = StateHelp
		return m, nil

	case views.ActionQuit:
		return m, tea.Quit
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
		return m.openSubMenu(views.MenuProton)

	case "3":
		return m.openSubMenu(views.MenuPerformance)

	case "4":
		return m.openSubMenu(views.MenuPrefix)

	case "5":
		return m.openSubMenu(views.MenuLogs)

	case "6":
		return m.openSubMenu(views.MenuSettings)

	// Direct hotkeys for fast power-user access
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

	case "x":
		m.Config.UseXalia = !m.Config.UseXalia
		_ = config.SaveConfig(m.GameDir, m.Config)
		return m, nil

	case "o":
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		m.OverridesView = views.NewOverridesView(pfxDir, m.Config.ProtonPath, m.ActiveOverrides, m.Config.DLLOverrides)
		m.State = StateOverrides
		return m, nil

	case "a":
		m.ProtonDBView = views.NewProtonDBView(m.ProtonDB, m.GameTitle, m.Config.AppID, m.Width, m.Height)
		m.State = StateProtonDB
		if m.ProtonDB != nil {
			return m, nil
		}
		return m, m.fetchProtonDB("")

	case "c":
		pfxDir := filepath.Join(m.GameDir, "proton-prefix")
		res, err := prefix.SafeCleanPrefixCustom(pfxDir, m.Config.ProtonPath, m.GameDir, m.Config.TargetExe, m.GameTitle, m.Config.BackupDir, m.Config.Filesystem.ExtraBackupPaths)
		if err != nil {
			m.StatusMessage = fmt.Sprintf("Clean failed: %v", err)
			m.PrefixCleaned = false
		} else if res.BackupDir != "" {
			m.StatusMessage = fmt.Sprintf("Prefix cleaned! Saved data preserved to: %s", res.BackupDir)
			m.PrefixCleaned = true
		} else {
			m.StatusMessage = "Prefix cleaned and reset successfully."
			m.PrefixCleaned = true
		}
		return m, nil

	case "m", "M":
		outputs := hardware.GetConnectedDisplayOutputs()
		options := append([]string{"auto"}, outputs...)
		curIdx := 0
		for i, o := range options {
			if strings.EqualFold(o, m.Config.GamescopeOutput) {
				curIdx = i
				break
			}
		}
		nextIdx := (curIdx + 1) % len(options)
		m.Config.GamescopeOutput = options[nextIdx]
		_ = config.SaveConfig(m.GameDir, m.Config)
		m.StatusMessage = fmt.Sprintf("Display output set to: %s", m.Config.GamescopeOutput)
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

	case "d", "D":
		p := quirks.GetActiveOrDetectedPreset(m.GameDir, m.Config.TargetExe, m.Config.AppID, m.Config.ProtonPath, m.Config)
		m.DetectedPreset = p
		m.PresetPickerView = views.NewPresetPickerView(p, m.GameTitle, m.Config, m.Width, m.Height)
		m.State = StatePresetPicker
		return m, nil

	case "L":
		m.Config.EnableLogging = !m.Config.EnableLogging
		_ = config.SaveConfig(m.GameDir, m.Config)
		if m.Config.EnableLogging {
			m.StatusMessage = "Session logging ENABLED (writes to .logs/ on launch)"
		} else {
			m.StatusMessage = "Session logging DISABLED"
		}
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
	case StatePresetPicker:
		if m.PresetPickerView != nil {
			return m.PresetPickerView.View()
		}
	case StateProtonDB:
		if m.ProtonDBView != nil {
			return m.ProtonDBView.View()
		}
	case StateSubMenu:
		if m.SubMenuView != nil {
			return m.SubMenuView.View()
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
