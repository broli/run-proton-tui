package diagnostics

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/hardware"
	"github.com/broli/run-proton-tui/internal/launcher"
)

// BugReportOptions defines options for generating a GitHub-ready issue report.
type BugReportOptions struct {
	Version string
	GameDir string
	Config  *config.GameConfig
}

// GenerateBugReport collects sanitized system diagnostics, hardware specs, and configuration
// following official Go bug reporting guidelines and formats a GitHub issue markdown document.
func GenerateBugReport(opts BugReportOptions) (string, string) {
	var sb strings.Builder

	rptVer := opts.Version
	if rptVer == "" {
		rptVer = "dev"
	}

	gameTitle := "Unknown Game"
	if opts.GameDir != "" {
		gameTitle = filepath.Base(opts.GameDir)
	}

	sb.WriteString("### Description of the Issue\n\n")
	sb.WriteString("<!-- A clear and concise description of what the problem is. -->\n\n")

	sb.WriteString("### Steps to Reproduce\n\n")
	sb.WriteString("1. cd into game directory: `" + SanitizePath(opts.GameDir) + "`\n")
	sb.WriteString("2. Run command: `rpt`\n")
	sb.WriteString("3. <!-- What action did you take next? (e.g. launched game, toggled Gamescope, changed runner) -->\n\n")

	sb.WriteString("### Expected Behavior\n\n")
	sb.WriteString("<!-- What did you expect to happen? -->\n\n")

	sb.WriteString("### Actual Behavior\n\n")
	sb.WriteString("<!-- What actually happened? Include error messages, exit codes, or crash symptoms. -->\n\n")

	sb.WriteString("### System & Environment Information\n\n")
	sb.WriteString("```text\n")
	sb.WriteString(fmt.Sprintf("rpt Version:      %s\n", rptVer))
	sb.WriteString(fmt.Sprintf("Go Runtime:       %s (%s/%s)\n", runtime.Version(), runtime.GOOS, runtime.GOARCH))

	if kernelOut, err := exec.Command("uname", "-sr").Output(); err == nil {
		sb.WriteString(fmt.Sprintf("Linux Kernel:     %s", string(kernelOut)))
	}

	desktop := os.Getenv("XDG_CURRENT_DESKTOP")
	if desktop == "" {
		desktop = os.Getenv("DESKTOP_SESSION")
	}
	sessionType := os.Getenv("XDG_SESSION_TYPE")
	if sessionType == "" {
		if os.Getenv("WAYLAND_DISPLAY") != "" {
			sessionType = "wayland"
		} else {
			sessionType = "x11"
		}
	}
	sb.WriteString(fmt.Sprintf("Desktop Session:  %s (%s)\n", desktop, sessionType))

	// Hardware Info
	gpuInfo, _ := hardware.DetectGPU()
	if gpuInfo != nil {
		sb.WriteString(fmt.Sprintf("GPU Prime-Run:    %v\n", gpuInfo.HasPrimeRun))
		sb.WriteString(fmt.Sprintf("GPU Hardware:     NVIDIA=%v, AMD=%v, Intel=%v\n", gpuInfo.HasNvidia, gpuInfo.HasAMD, gpuInfo.HasIntel))
		sb.WriteString(fmt.Sprintf("Display Outputs:  %s (Preferred: %s)\n", strings.Join(gpuInfo.ConnectedOutputs, ", "), gpuInfo.PreferredOutput))
	}
	sb.WriteString(fmt.Sprintf("Kernel NTSync:    %v (/dev/ntsync)\n", hardware.HasNTSync()))

	if topo, _ := hardware.DetectCPUTopology(); topo != nil {
		sb.WriteString(fmt.Sprintf("CPU Model:        %s\n", topo.ModelName))
		sb.WriteString(fmt.Sprintf("CPU Threads:      %d (Hybrid: %v, P-Cores: %s)\n", topo.TotalThreads, topo.IsHybrid, topo.PCoresMask))
	}
	sb.WriteString("```\n\n")

	// Active Configuration
	if opts.Config != nil {
		sb.WriteString("### Active Game Configuration (`.proton-config.toml`)\n\n")
		sb.WriteString("```toml\n")
		sb.WriteString(fmt.Sprintf("target_exe = %q\n", SanitizePath(opts.Config.TargetExe)))
		sb.WriteString(fmt.Sprintf("proton_path = %q\n", SanitizePath(opts.Config.ProtonPath)))
		sb.WriteString(fmt.Sprintf("app_id = %q\n", opts.Config.AppID))
		sb.WriteString(fmt.Sprintf("use_gamescope = %v\n", opts.Config.UseGamescope))
		if opts.Config.UseGamescope {
			sb.WriteString(fmt.Sprintf("gamescope_dimensions = \"%dx%d@%dHz -> %s\"\n",
				opts.Config.GamescopeWidth, opts.Config.GamescopeHeight, opts.Config.GamescopeRefresh, opts.Config.GamescopeOutput))
			sb.WriteString(fmt.Sprintf("gamescope_filter = %q\n", opts.Config.GamescopeFilter))
			sb.WriteString(fmt.Sprintf("gamescope_scaling = %q\n", opts.Config.GamescopeScaling))
		}
		sb.WriteString(fmt.Sprintf("use_prime_run = %v\n", opts.Config.UsePrimeRun))
		sb.WriteString(fmt.Sprintf("use_pcores = %v (mask: %q)\n", opts.Config.UsePCores, opts.Config.PCoresMask))
		sb.WriteString(fmt.Sprintf("manage_power = %v\n", opts.Config.ManagePower))
		sb.WriteString(fmt.Sprintf("use_xalia = %v\n", opts.Config.UseXalia))
		if opts.Config.PresetName != "" {
			sb.WriteString(fmt.Sprintf("preset_name = %q\n", opts.Config.PresetName))
		}
		if opts.Config.UmuID != "" {
			sb.WriteString(fmt.Sprintf("umu_id = %q\n", opts.Config.UmuID))
		}
		if len(opts.Config.DLLOverrides) > 0 {
			sb.WriteString("\n[dll_overrides]\n")
			for k, v := range opts.Config.DLLOverrides {
				sb.WriteString(fmt.Sprintf("%s = %q\n", k, v))
			}
		}
		sb.WriteString("```\n\n")
	}

	// Preflight Health Status
	if opts.GameDir != "" && opts.Config != nil {
		pfxDir := filepath.Join(opts.GameDir, "proton-prefix")
		report := RunPreflightCheck(opts.GameDir, opts.Config.TargetExe, pfxDir)
		sb.WriteString("### Health & Permissions Check\n\n")
		sb.WriteString(fmt.Sprintf("- Target Exists: `%v`\n", report.TargetExeExists))
		sb.WriteString(fmt.Sprintf("- Game Directory Writable: `%v`\n", report.GameDirWritable))
		sb.WriteString(fmt.Sprintf("- Prefix Directory Writable: `%v`\n", report.PrefixDirWritable))
		sb.WriteString(fmt.Sprintf("- Inside Prefix Hazard: `%v`\n", report.InsidePrefixHazard))
		if len(report.Issues) > 0 {
			sb.WriteString("- Issues Detected:\n")
			for _, iss := range report.Issues {
				sb.WriteString(fmt.Sprintf("  • %s\n", iss))
			}
		} else {
			sb.WriteString("- Issues Detected: None (Healthy)\n")
		}
		sb.WriteString("\n")
	}

	// Recent Log Tail
	if opts.GameDir != "" {
		pfxDir := filepath.Join(opts.GameDir, "proton-prefix")
		logs := FindRecentLogs(opts.GameDir, pfxDir)
		if len(logs) > 0 {
			latest := logs[0].Path
			tailLines := ReadLogTail(latest, 30)
			if len(tailLines) > 0 {
				sb.WriteString(fmt.Sprintf("### Log Excerpt (`%s`)\n\n", SanitizePath(latest)))
				sb.WriteString("```text\n")
				for _, l := range tailLines {
					sb.WriteString(SanitizePath(l) + "\n")
				}
				sb.WriteString("```\n\n")
			}
		}
	}

	sb.WriteString("### Privacy Confirmation\n\n")
	sb.WriteString("✓ All user home directory paths have been sanitized to `~`.\n")

	reportContent := sb.String()

	// Build GitHub URL
	issueTitle := fmt.Sprintf("[Bug]: issue launching %s", gameTitle)
	baseURL := "https://github.com/broli/run-proton-tui/issues/new"
	params := url.Values{}
	params.Set("title", issueTitle)
	params.Set("body", reportContent)
	targetURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	return reportContent, targetURL
}

// OpenBugReport generates the bug report, copies it to clipboard, and attempts to launch the browser.
func OpenBugReport(opts BugReportOptions) (string, error) {
	report, targetURL := GenerateBugReport(opts)
	_ = launcher.CopyToClipboard(report)
	err := launcher.OpenURL(targetURL)
	return report, err
}
