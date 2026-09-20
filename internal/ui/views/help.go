package views

import (
	"fmt"
	"strings"

	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// HelpView displays the in-depth manual and troubleshooting encyclopedia.
type HelpView struct {
	Viewport viewport.Model
	Ready    bool
}

// NewHelpView initializes the scrollable help manual.
func NewHelpView(width, height int) *HelpView {
	vp := viewport.New(width, height-6)
	vp.SetContent(renderHelpContent())

	return &HelpView{
		Viewport: vp,
		Ready:    true,
	}
}

// Update handles scroll keys within the help view.
func (h *HelpView) Update(msg tea.Msg) (*HelpView, tea.Cmd) {
	var cmd tea.Cmd
	h.Viewport, cmd = h.Viewport.Update(msg)
	return h, cmd
}

// View renders the help viewport and header.
func (h *HelpView) View() string {
	header := style.TitleStyle.Render("📖 rpt — In-Depth Manual & Troubleshooting Guide (Press [?] or [Esc] to Return)")
	footer := style.SubheaderStyle.Render(fmt.Sprintf("Scroll: ↑/k, ↓/j, PgUp, PgDown • Progress: %3.f%%", h.Viewport.ScrollPercent()*100))

	return lipgloss.JoinVertical(lipgloss.Left, header, h.Viewport.View(), footer)
}

func renderHelpContent() string {
	var sb strings.Builder

	section := func(title string) string {
		return style.HeaderStyle.Render("═══ " + title + " ═══\n")
	}
	sub := func(title string) string {
		return style.KeyStyle.Render("▶ " + title + "\n")
	}

	sb.WriteString(section("1. KEYBINDINGS REFERENCE"))
	sb.WriteString("  [Enter / 1] : Launch game immediately\n")
	sb.WriteString("  [2]         : Select Proton Runner (GE-Proton, CachyOS, Valve Experimental)\n")
	sb.WriteString("  [3]         : Select Executable (.exe picker with smart classification)\n")
	sb.WriteString("  [g]         : Toggle Gamescope Sandboxing (1080p fixed canvas)\n")
	sb.WriteString("  [p]         : Toggle CPU P-Core Pinning (taskset to threads 0-11)\n")
	sb.WriteString("  [v]         : Toggle GPU Runner (prime-run NVIDIA RTX 4060 vs Intel iGPU)\n")
	sb.WriteString("  [o]         : Configure Wine Registry DLL Overrides (UE4SS, ReShade, etc.)\n")
	sb.WriteString("  [a]         : Fetch and Apply ProtonDB Community Recommendations\n")
	sb.WriteString("  [x]         : Toggle Proton Xalia Accessibility Bridge\n")
	sb.WriteString("  [c]         : Clean / Reset Wine Prefix (with automatic savegame backup!)\n")
	sb.WriteString("  [h]         : Pre-flight Permissions & Environment Health Check\n")
	sb.WriteString("  [l]         : View Recent Session Logs & Crash Dumps\n")
	sb.WriteString("  [b]         : Toggle Opaque vs Transparent Terminal Backdrop\n")
	sb.WriteString("  [?] / [F1]  : Toggle this In-Depth Help Manual\n")
	sb.WriteString("  [q] / [Esc] : Exit rpt\n\n")

	sb.WriteString(section("2. STEAM EMULATORS & lsteamclient.dll=d"))
	sb.WriteString(sub("Why is lsteamclient disabled by default?"))
	sb.WriteString("  When running standalone games or repacks using emulators such as Goldberg,\n")
	sb.WriteString("  CODEX, Rune, Fairlight (FLT), or ALI213, the game folder includes custom\n")
	sb.WriteString("  steam_api.dll or steam_api64.dll wrappers.\n\n")
	sb.WriteString("  By default, Valve's Proton attempts to bridge Windows Steam API calls back to\n")
	sb.WriteString("  your native Linux Steam client via Wine's built-in 'lsteamclient.dll'.\n")
	sb.WriteString("  For non-Steam games, this triggers official Steam Store popups or causes\n")
	sb.WriteString("  access violation crashes. Setting WINEDLLOVERRIDES='lsteamclient=d' disables\n")
	sb.WriteString("  the Linux bridge completely, allowing local offline emulators to work seamlessly.\n\n")

	sb.WriteString(section("3. GAMESCOPE DECOUPLING & WAYLAND STABILITY"))
	sb.WriteString(sub("Why Gamescope must NOT run via prime-run on hybrid laptops"))
	sb.WriteString("  On hybrid Intel Iris Xe + NVIDIA RTX 4060 laptops running KDE Plasma Wayland:\n")
	sb.WriteString("  - KWin Wayland compositor runs on the Intel iGPU.\n")
	sb.WriteString("  - If Gamescope is launched with prime-run, it becomes an NVIDIA client streaming\n")
	sb.WriteString("    Wayland DMA-BUFs into Intel KWin. Under heavy 3D loads (DXVK swapchain changes),\n")
	sb.WriteString("    Intel DRM driver runs out of GEM execution buffer memory (ENOMEM),\n")
	sb.WriteString("    crashing the entire desktop session.\n\n")
	sb.WriteString("  - rpt's Proven Architecture: Gamescope runs natively as an Intel Wayland client,\n")
	sb.WriteString("    and prime-run is invoked INSIDE the sandbox wrapper strictly for the game binary.\n")
	sb.WriteString("    This eliminates fractional scaling distortion and prevents desktop crashes.\n\n")

	sb.WriteString(section("4. 2D LAUNCHERS VS 3D GAMES"))
	sb.WriteString(sub("Preventing NVAPI Double-Free Crashes"))
	sb.WriteString("  2D tools (like GRYPHLINK Games.exe, patchers, setup.exe, CEF/Qt5 WebEngine)\n")
	sb.WriteString("  crash with 'double free or corruption (!prev)' if forced to run via prime-run\n")
	sb.WriteString("  on Wayland. rpt's classification engine automatically assigns them to run\n")
	sb.WriteString("  on the host iGPU without Gamescope.\n\n")

	sb.WriteString(section("5. SAFE DIRECTORY LAYOUT & GUARDRAILS"))
	sb.WriteString(sub("Never install games inside drive_c!"))
	sb.WriteString("  Best Practice:\n")
	sb.WriteString("    ~/Games/EldenRing/               <-- Game installation root (all files live here)\n")
	sb.WriteString("    ~/Games/EldenRing/proton-prefix/ <-- Wine runtime sandbox (can be reset safely)\n\n")
	sb.WriteString("  Hazard Warning:\n")
	sb.WriteString("    If an installer places files inside ~/Games/EldenRing/proton-prefix/pfx/drive_c/,\n")
	sb.WriteString("    cleaning or resetting the prefix will DELETE your entire game! rpt automatically\n")
	sb.WriteString("    detects this condition and locks prefix cleaning with a safety error.\n\n")

	sb.WriteString(section("6. KERNEL FAST SYNCHRONIZATION (/dev/ntsync)"))
	sb.WriteString("  NTSync is the new Linux in-kernel NT synchronization driver available in CachyOS.\n")
	sb.WriteString("  It replaces futex-based Fsync/Esync with native kernel mutexes, eliminating\n")
	sb.WriteString("  high CPU context switching in multi-threaded Direct3D 11 & 12 games.\n")
	sb.WriteString("  rpt automatically detects /dev/ntsync and exports PROTON_USE_NTSYNC=1.\n\n")

	sb.WriteString(section("7. CPU PERFORMANCE CORE PINNING"))
	sb.WriteString("  Intel 12th/13th/14th Gen CPUs feature Performance (P) and Efficient (E) cores.\n")
	sb.WriteString("  Thread hopping to Gracemont E-cores causes micro-stutters and 1% low FPS drops.\n")
	sb.WriteString("  rpt reads /sys/devices/cpu_core/cpus (e.g. 0-11) and applies 'taskset -c 0-11'\n")
	sb.WriteString("  strictly confining the game to high-speed Golden Cove / Raptor Cove P-cores.\n\n")

	sb.WriteString(section("8. CLIPBOARD SYNC IN GAMESCOPE"))
	sb.WriteString("  Nested X11 compositors inside Wayland cannot normally share the clipboard.\n")
	sb.WriteString("  rpt integrates 'gamescope-clip-bridge' to stream host Wayland clipboard text\n")
	sb.WriteString("  (from wl-paste) directly into the game's X11 clipboard on DISPLAY=:1.\n")

	return sb.String()
}
