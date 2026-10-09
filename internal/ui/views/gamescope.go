package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var resolutionPresets = []struct {
	label  string
	width  int
	height int
}{
	{"Auto / Native (Blank - Host Negotiated)", 0, 0},
	{"1080p (1920x1080)", 1920, 1080},
	{"1440p (2560x1440)", 2560, 1440},
	{"4K UHD (3840x2160)", 3840, 2160},
	{"720p (1280x720)", 1280, 720},
	{"Deck (1280x800)", 1280, 800},
	{"900p (1600x900)", 1600, 900},
}

var refreshPresets = []struct {
	label string
	hz    int
}{
	{"Native / Untouched (0)", 0},
	{"60 Hz", 60},
	{"75 Hz", 75},
	{"120 Hz", 120},
	{"144 Hz", 144},
	{"165 Hz", 165},
	{"240 Hz", 240},
}

var windowModes = []string{"fullscreen", "borderless", "windowed"}
var filterPresets = []string{"", "linear", "fsr", "nis", "nearest", "pixel"}
var scalingPresets = []string{"", "fit", "fill", "stretch", "integer"}
var fpsLimits = []int{0, 30, 40, 60, 120}

type GamescopeField int

const (
	FieldOutput GamescopeField = iota
	FieldDetectSize
	FieldClearGeometry
	FieldResolution
	FieldRefresh
	FieldWindowMode
	FieldFilter
	FieldScaling
	FieldSharpness
	FieldAdaptiveSync
	FieldMangoApp
	FieldHDR
	FieldFPSLimit
	FieldResetDefaults
	fieldCount
)

// GamescopeView provides a fine-grained, interactive configuration editor
// for Gamescope sandboxing, display resolution, scaling, and feature flags.
type GamescopeView struct {
	Config        *config.GameConfig
	Cursor        int
	StatusMessage string
	Width         int
	Height        int
}

// NewGamescopeView creates an interactive settings view for Gamescope.
func NewGamescopeView(cfg *config.GameConfig, width, height int) *GamescopeView {
	return &GamescopeView{
		Config: cfg,
		Cursor: 0,
		Width:  width,
		Height: height,
	}
}

// Update processes navigation and value cycling in GamescopeView.
func (v *GamescopeView) Update(msg tea.Msg) (done bool, detected bool) {
	if keyMsg, ok := msg.(tea.KeyMsg); ok {
		k := keyMsg.String()
		switch k {
		case "esc", "q":
			return true, false

		case "up", "k":
			if v.Cursor > 0 {
				v.Cursor--
			} else {
				v.Cursor = int(fieldCount) - 1
			}

		case "down", "j":
			if v.Cursor < int(fieldCount)-1 {
				v.Cursor++
			} else {
				v.Cursor = 0
			}

		case "d", "D":
			v.DetectMonitorSize()
			return false, true

		case "u", "U", "x", "X", "backspace":
			v.ClearGeometry()
			return false, false

		case "r", "R":
			v.Config.ResetGamescopeToDefaults()
			v.StatusMessage = "✓ Gamescope reset to safe defaults (Auto Geometry, Linear, SDR, Untouched Hz)"
			return false, false

		case "enter", "space", " ", "right", "l":
			switch GamescopeField(v.Cursor) {
			case FieldOutput:
				v.cycleOutput(1)
			case FieldDetectSize:
				v.DetectMonitorSize()
				return false, true
			case FieldClearGeometry:
				v.ClearGeometry()
				return false, false
			case FieldResolution:
				v.cycleResolution(1)
			case FieldRefresh:
				v.cycleRefresh(1)
			case FieldWindowMode:
				v.cycleWindowMode(1)
			case FieldFilter:
				v.cycleFilter(1)
			case FieldScaling:
				v.cycleScaling(1)
			case FieldSharpness:
				v.adjSharpness(1)
			case FieldAdaptiveSync:
				v.Config.GamescopeAdaptiveSync = !v.Config.GamescopeAdaptiveSync
			case FieldMangoApp:
				v.Config.GamescopeMangoApp = !v.Config.GamescopeMangoApp
			case FieldHDR:
				v.Config.GamescopeHDR = !v.Config.GamescopeHDR
			case FieldFPSLimit:
				v.cycleFPSLimit(1)
			case FieldResetDefaults:
				v.Config.ResetGamescopeToDefaults()
				v.StatusMessage = "✓ Gamescope reset to safe defaults (Auto Geometry, Linear, SDR, Untouched Hz)"
			}

		case "left", "h":
			switch GamescopeField(v.Cursor) {
			case FieldOutput:
				v.cycleOutput(-1)
			case FieldDetectSize:
				v.DetectMonitorSize()
				return false, true
			case FieldClearGeometry:
				v.ClearGeometry()
				return false, false
			case FieldResolution:
				v.cycleResolution(-1)
			case FieldRefresh:
				v.cycleRefresh(-1)
			case FieldWindowMode:
				v.cycleWindowMode(-1)
			case FieldFilter:
				v.cycleFilter(-1)
			case FieldScaling:
				v.cycleScaling(-1)
			case FieldSharpness:
				v.adjSharpness(-1)
			case FieldAdaptiveSync:
				v.Config.GamescopeAdaptiveSync = !v.Config.GamescopeAdaptiveSync
			case FieldMangoApp:
				v.Config.GamescopeMangoApp = !v.Config.GamescopeMangoApp
			case FieldHDR:
				v.Config.GamescopeHDR = !v.Config.GamescopeHDR
			case FieldFPSLimit:
				v.cycleFPSLimit(-1)
			}

		case "+", "=":
			if GamescopeField(v.Cursor) == FieldSharpness {
				v.adjSharpness(1)
			}
		case "-", "_":
			if GamescopeField(v.Cursor) == FieldSharpness {
				v.adjSharpness(-1)
			}
		}
	}
	return false, false
}

// ClearGeometry resets Gamescope canvas width and height to 0 (Auto / Blank),
// allowing Gamescope to adaptively negotiate resolution with the host display.
func (v *GamescopeView) ClearGeometry() {
	v.Config.GamescopeWidth = 0
	v.Config.GamescopeHeight = 0
	v.Config.GeometryAutoDetected = false
	v.StatusMessage = "✓ Geometry unset to Auto / Blank (0x0 display negotiation)"
}

// DetectMonitorSize triggers DRM sysfs scanning for the active connector,
// updates width/height to native monitor dimensions, sets refresh to 0 (native/untouched),
// and marks GeometryAutoDetected so it saves with a documented comment in rpt.toml.
func (v *GamescopeView) DetectMonitorSize() {
	target := v.Config.GamescopeOutput
	w, h, err := hardware.DetectOutputResolution(target)
	v.Config.GamescopeWidth = w
	v.Config.GamescopeHeight = h
	v.Config.GamescopeRefresh = 0 // Untouched / Native
	v.Config.GeometryAutoDetected = true

	dispName := target
	if dispName == "" || strings.EqualFold(dispName, "auto") {
		dispName = "Auto-detected monitor"
	}
	if err != nil {
		v.StatusMessage = fmt.Sprintf("⚠️ Detection fallback (%s): %dx%d (Comment '# auto detected from TUI' will be saved)", dispName, w, h)
	} else {
		v.StatusMessage = fmt.Sprintf("✓ Detected %s: %dx%d (Comment '# auto detected from TUI' will be saved)", dispName, w, h)
	}
}

func (v *GamescopeView) cycleOutput(dir int) {
	outputs := hardware.GetConnectedDisplayOutputs()
	options := append([]string{"auto"}, outputs...)
	curIdx := 0
	for i, o := range options {
		if strings.EqualFold(o, v.Config.GamescopeOutput) {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + dir + len(options)) % len(options)
	v.Config.GamescopeOutput = options[nextIdx]
}

func (v *GamescopeView) cycleResolution(dir int) {
	curIdx := -1
	for i, r := range resolutionPresets {
		if r.width == v.Config.GamescopeWidth && r.height == v.Config.GamescopeHeight {
			curIdx = i
			break
		}
	}
	if curIdx == -1 {
		curIdx = 0
	}
	nextIdx := (curIdx + dir + len(resolutionPresets)) % len(resolutionPresets)
	v.Config.GamescopeWidth = resolutionPresets[nextIdx].width
	v.Config.GamescopeHeight = resolutionPresets[nextIdx].height
	if v.Config.GamescopeWidth == 0 && v.Config.GamescopeHeight == 0 {
		v.Config.GeometryAutoDetected = false
	}
}

func (v *GamescopeView) cycleRefresh(dir int) {
	curIdx := 0
	for i, r := range refreshPresets {
		if r.hz == v.Config.GamescopeRefresh {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + dir + len(refreshPresets)) % len(refreshPresets)
	v.Config.GamescopeRefresh = refreshPresets[nextIdx].hz
}

func (v *GamescopeView) cycleWindowMode(dir int) {
	cur := strings.ToLower(v.Config.GamescopeWindowMode)
	curIdx := 0
	for i, m := range windowModes {
		if m == cur {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + dir + len(windowModes)) % len(windowModes)
	v.Config.GamescopeWindowMode = windowModes[nextIdx]
}

func (v *GamescopeView) cycleFilter(dir int) {
	cur := strings.ToLower(v.Config.GamescopeFilter)
	curIdx := 0
	for i, f := range filterPresets {
		if f == cur {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + dir + len(filterPresets)) % len(filterPresets)
	v.Config.GamescopeFilter = filterPresets[nextIdx]
}

func (v *GamescopeView) cycleScaling(dir int) {
	cur := strings.ToLower(v.Config.GamescopeScaling)
	curIdx := 0
	for i, s := range scalingPresets {
		if s == cur {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + dir + len(scalingPresets)) % len(scalingPresets)
	v.Config.GamescopeScaling = scalingPresets[nextIdx]
}

func (v *GamescopeView) adjSharpness(delta int) {
	val := v.Config.GamescopeSharpness + delta
	if val < 0 {
		val = 0
	}
	if val > 20 {
		val = 20
	}
	v.Config.GamescopeSharpness = val
}

func (v *GamescopeView) cycleFPSLimit(dir int) {
	curIdx := 0
	for i, lim := range fpsLimits {
		if lim == v.Config.GamescopeFPSLimit {
			curIdx = i
			break
		}
	}
	nextIdx := (curIdx + dir + len(fpsLimits)) % len(fpsLimits)
	v.Config.GamescopeFPSLimit = fpsLimits[nextIdx]
}

// View renders the gamescope options dialog.
func (v *GamescopeView) View() string {
	contentWidth := style.ClampWidth(v.Width, 10, 70, 96)

	header := lipgloss.NewStyle().
		Bold(true).
		Foreground(style.ColorPrimary).
		Render("🎯 GAMESCOPE SANDBOX & UPSCALING CONFIGURATION")

	subHeader := lipgloss.NewStyle().
		Foreground(style.ColorMuted).
		Render("Configure nested resolution, native refresh rate, FSR upscaling, and DRM output")

	divider := lipgloss.NewStyle().
		Foreground(style.ColorBorder).
		Render(strings.Repeat("─", contentWidth))

	var b strings.Builder
	b.WriteString(header + "\n")
	b.WriteString(subHeader + "\n")
	b.WriteString(divider + "\n\n")

	if v.StatusMessage != "" {
		b.WriteString(style.BadgeSuccess.Render(" ℹ "+v.StatusMessage) + "\n\n")
	}

	// Format values
	dispVal := v.Config.GamescopeOutput
	if dispVal == "" || strings.EqualFold(dispVal, "auto") {
		dispVal = "Auto"
	}

	resVal := "Auto / Native (Host Negotiated)"
	if v.Config.GamescopeWidth > 0 && v.Config.GamescopeHeight > 0 {
		resVal = fmt.Sprintf("%dx%d", v.Config.GamescopeWidth, v.Config.GamescopeHeight)
	}
	for _, r := range resolutionPresets {
		if r.width == v.Config.GamescopeWidth && r.height == v.Config.GamescopeHeight {
			resVal = r.label
			break
		}
	}

	refVal := "Native / Untouched (0)"
	if v.Config.GamescopeRefresh > 0 {
		refVal = fmt.Sprintf("%d Hz (Forced)", v.Config.GamescopeRefresh)
	}

	winVal := "Fullscreen (-f)"
	if strings.EqualFold(v.Config.GamescopeWindowMode, "borderless") {
		winVal = "Borderless Window (-b)"
	} else if strings.EqualFold(v.Config.GamescopeWindowMode, "windowed") {
		winVal = "Windowed"
	}

	filterVal := "Auto / Default (blank)"
	if strings.EqualFold(v.Config.GamescopeFilter, "linear") {
		filterVal = "Linear (Default / Clean)"
	} else if strings.EqualFold(v.Config.GamescopeFilter, "fsr") {
		filterVal = "AMD FidelityFX FSR 1.0"
	} else if strings.EqualFold(v.Config.GamescopeFilter, "nis") {
		filterVal = "NVIDIA Image Scaling (NIS)"
	} else if v.Config.GamescopeFilter != "" {
		filterVal = style.Capitalize(v.Config.GamescopeFilter)
	}

	scaleVal := "Auto / Default (blank)"
	if strings.EqualFold(v.Config.GamescopeScaling, "fit") {
		scaleVal = "Fit (Aspect Ratio Preserved)"
	} else if v.Config.GamescopeScaling != "" {
		scaleVal = style.Capitalize(v.Config.GamescopeScaling)
	}

	sharpVal := fmt.Sprintf("%d / 20 (0=Max, 20=Min)", v.Config.GamescopeSharpness)
	if v.Config.GamescopeSharpness == 0 {
		sharpVal = "0 (Max Sharpness / Default)"
	}

	fpsVal := "Disabled (Unlimited)"
	if v.Config.GamescopeFPSLimit > 0 {
		fpsVal = fmt.Sprintf("%d FPS Limit", v.Config.GamescopeFPSLimit)
	}

	boolBadge := func(on bool) string {
		if on {
			return style.BadgeSuccess.Render("ENABLED")
		}
		return style.BadgeMuted.Render("Disabled")
	}

	items := []struct {
		field GamescopeField
		label string
		val   string
	}{
		{FieldOutput, "Target Display Monitor", dispVal},
		{FieldDetectSize, "[⚡ Detect Monitor Size]", "Auto-detect native resolution & record comment"},
		{FieldClearGeometry, "[✖ Leave Geometry Blank]", "Reset width/height to 0 (Auto / Host Negotiated)"},
		{FieldResolution, "Output Resolution (-W / -H)", resVal},
		{FieldRefresh, "Refresh Rate Override (-r)", refVal},
		{FieldWindowMode, "Window Display Mode", winVal},
		{FieldFilter, "Upscaling Filter (-F)", filterVal},
		{FieldScaling, "Scaling Strategy (-S)", scaleVal},
		{FieldSharpness, "Upscaler Sharpness (--sharpness)", sharpVal},
		{FieldAdaptiveSync, "Adaptive Sync / VRR (--adaptive-sync)", boolBadge(v.Config.GamescopeAdaptiveSync)},
		{FieldMangoApp, "MangoHud Overlay (--mangoapp)", boolBadge(v.Config.GamescopeMangoApp)},
		{FieldHDR, "HDR Output (--hdr-enabled)", boolBadge(v.Config.GamescopeHDR)},
		{FieldFPSLimit, "Framerate Limit (--framerate-limit)", fpsVal},
		{FieldResetDefaults, "[🔄 Reset to Safe Defaults]", "Restore safe 1080p, Untouched Hz, Linear, SDR defaults"},
	}

	for i, it := range items {
		prefix := "   "
		lblStyle := lipgloss.NewStyle().Foreground(style.ColorText)
		if i == v.Cursor {
			prefix = " > "
			lblStyle = lipgloss.NewStyle().Bold(true).Foreground(style.ColorHighlight)
		}

		b.WriteString(fmt.Sprintf("%s%-36s : %s\n", prefix, lblStyle.Render(it.label), it.val))
	}

	b.WriteString("\n" + divider + "\n")
	navHelp := fmt.Sprintf("%s    %s    %s    %s    %s",
		lipgloss.NewStyle().Foreground(style.ColorHighlight).Render("[↑/↓] Navigate"),
		lipgloss.NewStyle().Foreground(style.ColorSecondary).Render("[Enter/Space/←/→] Select"),
		lipgloss.NewStyle().Foreground(style.ColorSuccess).Render("[d] Detect"),
		lipgloss.NewStyle().Foreground(style.ColorWarning).Render("[u] Blank Geometry"),
		lipgloss.NewStyle().Foreground(style.ColorMuted).Render("[r] Defaults"),
	)
	exitHelp := lipgloss.NewStyle().Foreground(style.ColorMuted).Render("    [Esc/q] Return")
	b.WriteString(navHelp + exitHelp)

	return lipgloss.NewStyle().
		BorderStyle(lipgloss.RoundedBorder()).
		BorderForeground(style.ColorPrimary).
		Padding(1, 2).
		Width(contentWidth + 4).
		Render(b.String())
}
