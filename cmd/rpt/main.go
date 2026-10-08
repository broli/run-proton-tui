package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/hooks"
	"github.com/broli/run-proton-tui/internal/integrations"
	"github.com/broli/run-proton-tui/internal/launcher"
	"github.com/broli/run-proton-tui/internal/prefix"
	"github.com/broli/run-proton-tui/internal/proton"
	"github.com/broli/run-proton-tui/internal/quirks"
	"github.com/broli/run-proton-tui/internal/runner"
	"github.com/broli/run-proton-tui/internal/ui"
	"github.com/broli/run-proton-tui/internal/ui/style"
	"github.com/broli/run-proton-tui/internal/ui/views"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	version = "0.9.0"
)

func main() {
	flagNow := flag.Bool("now", false, "Launch immediately with saved or auto-detected settings (skip TUI)")
	flagNoTUI := flag.Bool("no-tui", false, "Alias for --now")
	flagYes := flag.Bool("y", false, "Alias for --now")
	flagClean := flag.Bool("clean", false, "Clean / reset Wine prefix before launch (with automatic save backup)")
	flagCleanShort := flag.Bool("c", false, "Alias for --clean")
	flagCleanZero := flag.Bool("clean-zero", false, "Launch in pure upstream Proton baseline with zero extra flags or wrappers (Gate 1)")
	flagVanilla := flag.Bool("vanilla", false, "Alias for --clean-zero")
	flagLog := flag.Bool("log", false, "Enable verbose Proton & DXVK logging to .logs/")
	flagLogShort := flag.Bool("v", false, "Alias for --log")
	flagDiagnostics := flag.Bool("diagnostics", false, "Run pre-flight health & permissions check and exit")
	flagDiag := flag.Bool("diag", false, "Alias for --diagnostics")
	flagDiagShort := flag.Bool("d", false, "Alias for --diagnostics")
	flagVersion := flag.Bool("version", false, "Show version information")
	flagDumpSpec := flag.Bool("dump-spec", false, "Export system and game details for an AI assistant (like ChatGPT or Claude) to configure your game")
	flagHelpDump := flag.Bool("helpdump", false, "Alias for --dump-spec")
	flagCreateDesktop := flag.Bool("create-desktop", false, "Generate an XDG .desktop application launcher icon for this game")
	flagDesktopDir := flag.String("desktop-dir", "applications", "Destination for .desktop shortcut (applications, desktop, local)")
	flagInspectHooks := flag.Bool("inspect-hooks", false, "Inspect resolved lifecycle hooks, search order, and exported environment variables")
	flagHooks := flag.Bool("hooks", false, "Alias for --inspect-hooks")
	flagReportProtonDB := flag.Bool("report-protondb", false, "Generate a ProtonDB compatibility report markdown and copy to clipboard")
	flagSubmitQuirk := flag.Bool("submit-quirk", false, "Submit game quirk profile to broli/run-proton-tui (gh -> git -> web)")
	flagBugReport := flag.Bool("bug-report", false, "Generate a sanitized system diagnostic bug report for GitHub issues")
	flagBug := flag.Bool("bug", false, "Alias for --bug-report")

	// Toggles
	flagGamescope := flag.String("gamescope", "", "Force Gamescope on/off (true/false)")
	flagPCores := flag.String("pcores", "", "Force CPU P-Core pinning on/off (true/false)")
	flagXalia := flag.String("xalia", "", "Force Proton Xalia on/off (true/false)")
	flagProfile := flag.String("profile", "", "Select specific game profile from rpt.toml (e.g. game, launcher)")
	flagDryRun := flag.Bool("dry-run", false, "Print resolved launch command without executing")
	flagPrintCmd := flag.Bool("print-cmd", false, "Alias for --dry-run")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "rpt (Run Proton TUI) v%s — Modular Linux Game Launcher Helper\n\n", version)
		fmt.Fprintf(os.Stderr, "Recommended Workflow:\n")
		fmt.Fprintf(os.Stderr, "  cd into your game's root directory and simply run 'rpt'.\n")
		fmt.Fprintf(os.Stderr, "  rpt automatically discovers executables, separates 3D engines from 2D utilities,\n")
		fmt.Fprintf(os.Stderr, "  manages isolated prefixes (./proton-prefix/), and loads game quirks.\n")
		fmt.Fprintf(os.Stderr, "  (You do NOT need to specify the .exe manually unless selecting a specific tool).\n\n")
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  rpt [flags] [optional-game.exe] [-- extra-game-args...]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  cd ~/Games/Endfield && rpt      # Interactive TUI in game directory (recommended)\n")
		fmt.Fprintf(os.Stderr, "  rpt --now                       # Quick launch with saved/auto-detected settings\n")
		fmt.Fprintf(os.Stderr, "  rpt Setup.exe                   # Switch to specific installer/tool and open TUI\n")
		fmt.Fprintf(os.Stderr, "  rpt --profile launcher --now    # Launch specific profile headlessly\n")
		fmt.Fprintf(os.Stderr, "  rpt --dry-run                   # Inspect resolved launch command without executing\n")
		fmt.Fprintf(os.Stderr, "  rpt --clean-zero --now          # Test pure upstream Proton baseline (zero extra flags)\n")
		fmt.Fprintf(os.Stderr, "  rpt --clean --now               # Safe prefix reset followed by instant launch\n")
		fmt.Fprintf(os.Stderr, "  rpt --create-desktop            # Create 1-click desktop/applications menu shortcut\n")
		fmt.Fprintf(os.Stderr, "  rpt --dump-spec                 # Output system details and hook guidance for AI assistants\n")
		fmt.Fprintf(os.Stderr, "  rpt --report-protondb           # Generate standardized ProtonDB Markdown report & copy to clipboard\n")
		fmt.Fprintf(os.Stderr, "  rpt --submit-quirk              # Submit game quirks to broli/run-proton-tui under MIT License\n")
		fmt.Fprintf(os.Stderr, "  rpt --bug-report                # Generate sanitized bug report, copy to clipboard & open GitHub\n")
	}

	flag.Parse()

	if *flagDumpSpec || *flagHelpDump {
		specBytes, err := diagnostics.GenerateSpecDump(version)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error generating specification dump: %v\n", err)
			os.Exit(1)
		}
		fmt.Println(string(specBytes))
		os.Exit(0)
	}

	if *flagVersion {
		fmt.Printf("rpt v%s\n", version)
		os.Exit(0)
	}

	gameDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	skipTUI := *flagNow || *flagNoTUI || *flagYes || !isatty.IsTerminal(os.Stdin.Fd())

	// 1. Load configuration from rpt.toml
	cfgFile, err := config.LoadConfigFile(gameDir)
	if err != nil {
		if os.IsNotExist(err) {
			if skipTUI && !(*flagDryRun || *flagPrintCmd) {
				// User attempted headless launch without an rpt.toml config
				errBox := lipgloss.NewStyle().
					BorderStyle(lipgloss.DoubleBorder()).
					BorderForeground(style.ColorDanger).
					Padding(1, 2).
					Width(78)

				titleStyle := lipgloss.NewStyle().Bold(true).Foreground(style.ColorDanger)
				bodyStyle := lipgloss.NewStyle().Foreground(style.ColorText)
				cmdStyle := lipgloss.NewStyle().Bold(true).Foreground(style.ColorPrimary)
				subStyle := lipgloss.NewStyle().Foreground(style.ColorMuted)

				msg := fmt.Sprintf(
					"%s\n\n"+
						"%s\n\n"+
						"To configure this game, run interactive mode in this directory:\n"+
						"  %s\n\n"+
						"%s\n\n"+
						"This allows rpt to detect executables, select your Proton runner,\n"+
						"and save a tailored 'rpt.toml' profile before running headlessly.\n",
					titleStyle.Render("No Configuration Found (rpt.toml)"),
					bodyStyle.Render("Headless launch (--now) requires an existing game profile in rpt.toml."),
					cmdStyle.Render(fmt.Sprintf("$ cd %q && rpt", gameDir)),
					subStyle.Render("(Generic syntax: cd /path/to/game/folder && rpt)"),
				)

				fmt.Fprintln(os.Stderr, errBox.Render(msg))
				os.Exit(1)
			}
			cfgFile = config.NewConfigFileWithDefault()
			config.ApplyHardwareSafeStandards(cfgFile.GetActiveProfile())
		} else {
			fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
			os.Exit(1)
		}
	}

	if *flagProfile != "" {
		if err := cfgFile.SetActiveProfile(*flagProfile); err != nil {
			fmt.Fprintf(os.Stderr, "Error selecting profile %q: %v\n", *flagProfile, err)
			fmt.Fprintf(os.Stderr, "Available profiles in rpt.toml: %s\n", strings.Join(cfgFile.ProfileNames(), ", "))
			os.Exit(1)
		}
	}

	cfg := cfgFile.GetActiveProfile()

	// Positional arguments: check for custom executable
	args := flag.Args()
	if len(args) > 0 {
		candidate := args[0]
		if strings.HasSuffix(strings.ToLower(candidate), ".exe") {
			cfg.TargetExe = candidate
			args = args[1:]
		}
	}

	// Auto-resolve target executable if not specified
	if cfg.TargetExe == "" {
		exes := views.DiscoverExecutables(gameDir)
		if len(exes) > 0 {
			cfg.TargetExe = exes[0].RelativePath
		}
	}

	// Auto-resolve proton runner if not specified
	if cfg.ProtonPath == "" {
		runners, _ := proton.DiscoverRunners()
		if len(runners) > 0 {
			cfg.ProtonPath = runners[0].Path
		}
	}

	// Auto-detect quirks preset if not yet configured
	if cfg.PresetName == "" && cfg.TargetExe != "" {
		if p := quirks.DetectQuirks(gameDir, cfg.TargetExe, cfg.AppID, cfg.ProtonPath); p != nil {
			p.ApplyToConfig(cfg, false)
		}
	}

	if *flagDryRun || *flagPrintCmd {
		fmt.Println("================================================================================")
		fmt.Printf("🚀 rpt Launch Command Dry-Run (Profile: %s)\n", cfgFile.ActiveProfile)
		fmt.Println("================================================================================")
		fmt.Printf("• Game Directory:   %s\n", gameDir)
		fmt.Printf("• Target Binary:    %s\n", cfg.TargetExe)
		fmt.Printf("• Proton Runner:    %s (%s)\n", filepath.Base(cfg.ProtonPath), cfg.ProtonPath)
		if cfg.UseGamescope {
			refStr := "Native"
			if cfg.GamescopeRefresh > 0 {
				refStr = fmt.Sprintf("%dHz", cfg.GamescopeRefresh)
			}
			geomStr := "Auto (Host Native)"
			if cfg.GamescopeWidth > 0 && cfg.GamescopeHeight > 0 {
				geomStr = fmt.Sprintf("%dx%d", cfg.GamescopeWidth, cfg.GamescopeHeight)
			}
			fmt.Printf("• Gamescope Mode:   Active (Output: %s, Geometry: %s, Refresh: %s, Window: %s)\n",
				cfg.GamescopeOutput, geomStr, refStr, cfg.GamescopeWindowMode)
		} else {
			fmt.Println("• Gamescope Mode:   Disabled (Native Window)")
		}
		if cfg.UsePCores {
			fmt.Printf("• CPU Affinity:     Pinned to P-Cores (%s)\n", cfg.PCoresMask)
		}
		if cfg.UsePrimeRun {
			fmt.Println("• GPU Offload:      prime-run (Dedicated GPU)")
		}
		fmt.Println("\n--- Injected Environment Overrides ---")
		if len(cfg.EnvVars) > 0 {
			for k, v := range cfg.EnvVars {
				fmt.Printf("  export %s=%q\n", k, v)
			}
		} else {
			fmt.Println("  (None)")
		}

		prefixCmd := ""
		if cfg.UsePrimeRun {
			prefixCmd = "prime-run "
		}
		if cfg.UsePCores && cfg.PCoresMask != "" {
			prefixCmd = fmt.Sprintf("taskset -c %s %s", cfg.PCoresMask, prefixCmd)
		}
		innerCmd := fmt.Sprintf("%s%q waitforexitandrun ./%q", prefixCmd, cfg.ProtonPath, cfg.TargetExe)
		if cfg.UseGamescope {
			gsArgs := runner.BuildGamescopeArgs(cfg)
			if len(gsArgs) > 0 {
				fmt.Printf("\n--- Resolved Execution Command ---\ngamescope %s -- %s\n\n", strings.Join(gsArgs, " "), innerCmd)
			} else {
				fmt.Printf("\n--- Resolved Execution Command ---\ngamescope -- %s\n\n", innerCmd)
			}
		} else {
			fmt.Printf("\n--- Resolved Execution Command ---\n%s\n\n", innerCmd)
		}
		os.Exit(0)
	}

	if *flagBugReport || *flagBug {
		report, err := diagnostics.OpenBugReport(diagnostics.BugReportOptions{
			Version: version,
			GameDir: gameDir,
			Config:  cfg,
		})
		fmt.Println(report)
		if err != nil {
			fmt.Printf("\n[✓] Bug report copied to clipboard. (Could not open browser automatically: %v)\n", err)
		} else {
			fmt.Println("\n[✓] Bug report copied to clipboard and GitHub issue page opened in browser.")
		}
		os.Exit(0)
	}

	// Apply CLI overrides to configuration
	if *flagCreateDesktop {
		loc := launcher.LocationApplications
		switch strings.ToLower(*flagDesktopDir) {
		case "desktop":
			loc = launcher.LocationDesktop
		case "local", "here", ".":
			loc = launcher.LocationLocal
		}
		gameTitle := filepath.Base(gameDir)
		if cfg.PresetName != "" {
			gameTitle = cfg.PresetName
		}
		opts := launcher.ShortcutOptions{
			GameTitle: gameTitle,
			GameDir:   gameDir,
			TargetExe: cfg.TargetExe,
			Location:  loc,
		}
		path, err := launcher.CreateDesktopShortcut(opts)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error creating desktop shortcut: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("[✓] Desktop shortcut created successfully: %s\n", path)
		os.Exit(0)
	}

	if *flagInspectHooks || *flagHooks {
		pfxDir := filepath.Join(gameDir, "proton-prefix")
		hookOpts := hooks.HookOptions{
			GameDir:        gameDir,
			PrefixDir:      pfxDir,
			TargetExe:      cfg.TargetExe,
			ProtonPath:     cfg.ProtonPath,
			ConfigHookPath: cfg.PreLaunchHook,
			PreLaunchHook:  cfg.PreLaunchHook,
			PostExitHook:   cfg.PostExitHook,
			ExtraHookDirs:  cfg.HookDirs,
		}
		if cfg.UseGamescope {
			hookOpts.GamescopeDisplay = ":1"
		}
		report := hooks.InspectAllHooks(hookOpts)

		fmt.Println("================================================================================")
		fmt.Println("🔗 rpt Lifecycle Hook Inspector & Presence Report")
		fmt.Println("================================================================================")

		preStatus := "None detected (standard launch)"
		if report.PreLaunch.Resolved != "" {
			preStatus = report.PreLaunch.Resolved
		}
		fmt.Printf("• Pre-Launch Hook: %s\n", preStatus)

		postStatus := "None detected (no cleanup script)"
		if report.PostExit.Resolved != "" {
			postStatus = report.PostExit.Resolved
		}
		fmt.Printf("• Post-Exit Hook:  %s\n\n", postStatus)

		fmt.Println("--- Pre-Launch Search Order Hierarchy ---")
		for _, c := range report.PreLaunch.Candidates {
			marker := "[ ]"
			if c.Exists {
				if c.Selected {
					marker = "[✓ ACTIVE]"
				} else {
					marker = "[• SHADOW]"
				}
			}
			fmt.Printf("  %-11s %s (%s)\n", marker, c.Path, c.Source)
		}
		fmt.Println()

		fmt.Println("--- Post-Exit Search Order Hierarchy ---")
		for _, c := range report.PostExit.Candidates {
			marker := "[ ]"
			if c.Exists {
				if c.Selected {
					marker = "[✓ ACTIVE]"
				} else {
					marker = "[• SHADOW]"
				}
			}
			fmt.Printf("  %-11s %s (%s)\n", marker, c.Path, c.Source)
		}
		fmt.Println()

		fmt.Println("--- Injected Environment Variables ($RPT_* Cheatsheet) ---")
		for _, env := range report.EnvVars {
			fmt.Printf("  $%-22s = %s\n", env.Name, env.CurrentVal)
			fmt.Printf("    └─ %s\n", env.Description)
		}
		fmt.Println()

		if report.PreLaunch.Resolved != "" {
			if data, err := os.ReadFile(report.PreLaunch.Resolved); err == nil {
				fmt.Printf("--- Pre-Launch Script Content (%s) ---\n", report.PreLaunch.Resolved)
				if hl, err := hooks.HighlightBash(string(data)); err == nil {
					fmt.Print(hl)
				} else {
					fmt.Print(string(data))
				}
				fmt.Println()
			}
		}

		os.Exit(0)
	}

	if *flagClean || *flagCleanShort {
		pfxDir := filepath.Join(gameDir, "proton-prefix")
		res, err := prefix.SafeCleanPrefixCustom(pfxDir, cfg.ProtonPath, gameDir, cfg.TargetExe, filepath.Base(gameDir), cfg.BackupDir, cfg.Filesystem.ExtraBackupPaths)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Prefix clean aborted: %v\n", err)
			os.Exit(1)
		}
		if res.BackupDir != "" {
			fmt.Printf("[✓] Prefix reset! Save games backed up to: %s\n", res.BackupDir)
		} else {
			fmt.Println("[✓] Prefix cleaned and reset successfully.")
		}
	}

	if *flagCleanZero || *flagVanilla {
		cfg.ResetToCleanZero()
	}
	if *flagLog || *flagLogShort {
		cfg.EnableLogging = true
	}
	if *flagGamescope != "" {
		cfg.UseGamescope = (*flagGamescope == "true" || *flagGamescope == "1")
	}
	if *flagPCores != "" {
		cfg.UsePCores = (*flagPCores == "true" || *flagPCores == "1")
	}
	if *flagXalia != "" {
		cfg.UseXalia = (*flagXalia == "true" || *flagXalia == "1")
	}

	pfxDir := filepath.Join(gameDir, "proton-prefix")

	// Diagnostics flag
	if *flagDiagnostics || *flagDiag || *flagDiagShort {
		report := diagnostics.RunPreflightCheck(gameDir, cfg.TargetExe, pfxDir)
		fmt.Printf("=== Health Report for %s ===\n", filepath.Base(gameDir))
		fmt.Printf("Target Exe: %s (Exists: %v, Auto-Fixed +x: %v)\n", cfg.TargetExe, report.TargetExeExists, report.TargetExeFixed)
		fmt.Printf("Inside Prefix Hazard: %v\n", report.InsidePrefixHazard)
		fmt.Printf("Game Dir Writable: %v\n", report.GameDirWritable)
		fmt.Printf("Prefix Dir Writable: %v\n", report.PrefixDirWritable)
		for _, issue := range report.Issues {
			fmt.Printf("Issue: %s\n", issue)
		}
		os.Exit(0)
	}

	// 2. Pre-flight health check (auto-adds +x if missing)
	diagnostics.RunPreflightCheck(gameDir, cfg.TargetExe, pfxDir)

	if *flagReportProtonDB {
		history, _ := runner.ReadSessionHistory(gameDir)
		report := integrations.GenerateProtonDBReport(integrations.ExportOptions{
			AppID:     cfg.AppID,
			GameTitle: filepath.Base(gameDir),
			GameDir:   gameDir,
			Config:    cfg,
			History:   history,
		})
		fmt.Println(report)
		if err := launcher.CopyToClipboard(report); err == nil {
			fmt.Println("\n[✓] Report copied to system clipboard!")
		}
		targetURL := fmt.Sprintf("https://www.protondb.com/app/%s", cfg.AppID)
		if !cfg.HasValidAppID() {
			targetURL = "https://www.protondb.com/contribute"
		}
		_ = launcher.OpenURL(targetURL)
		fmt.Printf("[✓] Opened %s in your default browser.\n", targetURL)
		os.Exit(0)
	}

	if *flagSubmitQuirk {
		fmt.Println(quirks.MITConsentNotice)
		fmt.Print("\nDo you accept the MIT License terms to contribute this quirk? [y/N]: ")
		var resp string
		fmt.Scanln(&resp)
		if strings.ToLower(strings.TrimSpace(resp)) != "y" && strings.ToLower(strings.TrimSpace(resp)) != "yes" {
			fmt.Println("Aborted: MIT License terms must be accepted to contribute quirks.")
			os.Exit(1)
		}

		res, err := quirks.SubmitQuirk(gameDir, cfg, true)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Submission failed: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("\n[%s] %s\n", res.Method, res.Message)
		if res.URL != "" {
			fmt.Printf("URL: %s\n", res.URL)
		}
		os.Exit(0)
	}

	if !skipTUI {
		// Run Interactive TUI
		m, err := ui.NewModel(gameDir, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize TUI: %v\n", err)
			os.Exit(1)
		}
		m.Version = version

		p := tea.NewProgram(m, tea.WithAltScreen())
		finalModel, err := p.Run()
		if err != nil {
			fmt.Fprintf(os.Stderr, "TUI error: %v\n", err)
			os.Exit(1)
		}

		appModel, ok := finalModel.(*ui.Model)
		if !ok || !appModel.ShouldLaunch {
			// User exited without launching
			os.Exit(0)
		}
		cfg = appModel.Config
		cfgFile = appModel.ConfigFile
		_ = config.SaveConfigFile(gameDir, cfgFile)
	}

	// Guardrail: Verify TargetExe exists in gameDir before launching
	targetPath := filepath.Join(gameDir, cfg.TargetExe)
	if _, err := os.Stat(targetPath); os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "\n❌ Executable Not Found: %q does not exist in:\n   %s\n\n", cfg.TargetExe, gameDir)

		var availableExes []string
		if entries, err := os.ReadDir(gameDir); err == nil {
			for _, entry := range entries {
				if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".exe") {
					availableExes = append(availableExes, entry.Name())
				}
			}
		}

		if len(availableExes) > 0 {
			fmt.Fprintf(os.Stderr, "   Available executable(s) in this directory:\n")
			for _, ex := range availableExes {
				fmt.Fprintf(os.Stderr, "     • %s\n", ex)
			}
			fmt.Fprintf(os.Stderr, "\n   To launch, run: rpt %s\n\n", availableExes[0])
		} else {
			fmt.Fprintf(os.Stderr, "   No .exe files found here. Make sure you are in the intended game directory.\n\n")
		}
		os.Exit(1)
	}

	// 3. Execute Game Runner - Linux Boot Style status lines
	okStyle := lipgloss.NewStyle().Bold(true).Foreground(style.ColorSuccess)
	fmt.Printf("\n%s Target executable verified: %s\n", okStyle.Render("[ OK ]"), cfg.TargetExe)
	if preHook := hooks.ResolveHook(gameDir, hooks.PreLaunch, cfg.PreLaunchHook, cfg.HookDirs...); preHook != "" {
		fmt.Printf("%s Pre-launch lifecycle hook found: %s\n", okStyle.Render("[ OK ]"), preHook)
	}
	if cfg.UsePCores {
		fmt.Printf("%s CPU affinity pinned to P-Cores (Threads %s)\n", okStyle.Render("[ OK ]"), cfg.PCoresMask)
	}
	if cfg.UseGamescope {
		refStr := "Native"
		if cfg.GamescopeRefresh > 0 {
			refStr = fmt.Sprintf("%dHz", cfg.GamescopeRefresh)
		}
		fmt.Printf("%s Gamescope sandboxing active (%dx%d @ %s -> %s)\n",
			okStyle.Render("[ OK ]"), cfg.GamescopeWidth, cfg.GamescopeHeight, refStr, cfg.GamescopeOutput)
	}
	if cfg.UsePrimeRun {
		fmt.Printf("%s Dedicated GPU offload active (prime-run)\n", okStyle.Render("[ OK ]"))
	}
	fmt.Printf("%s Launching Proton runner (%s)...\n\n", okStyle.Render("[ OK ]"), filepath.Base(cfg.ProtonPath))

	opts := runner.LaunchOptions{
		GameDir:    gameDir,
		Config:     cfg,
		ExtraArgs:  args,
		StdoutPipe: os.Stdout,
		StderrPipe: os.Stderr,
	}

	// Trap SIGINT (Ctrl+C) and SIGTERM so user aborts can be intercepted cleanly,
	// child processes safely terminated, prefix flushed, and post-session diagnostics displayed.
	ctx, cancel := context.WithCancel(context.Background())
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		select {
		case sig, ok := <-sigChan:
			if ok && sig != nil {
				cancel()
			}
		case <-ctx.Done():
		}
	}()
	defer func() {
		signal.Stop(sigChan)
		cancel()
	}()

	result, err := runner.RunGame(ctx, opts)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Execution error: %v\n", err)
		os.Exit(1)
	}

	// Record local session telemetry (strictly no-op if cfg.EnableLocalTelemetry is false)
	_ = runner.RecordSession(gameDir, cfg, result, "")

	// 4. Progressive Fallback, Crash & Stalled Session Interception
	if result.AbortedByUser || result.CrashDetected {
		helpDir := filepath.Join(gameDir, ".logs")
		_ = os.MkdirAll(helpDir, 0755)
		absHelpFilePath := filepath.Join(helpDir, "ask-ai-help.txt")

		insights := runner.AnalyzeSessionLog(result.LogFile, cfg, result.ExitCode)

		// Generate and write AI Assistant Helper package to log file only (hiding complexity from user)
		logTail := runner.ReadLogTail(result.LogFile, 35)
		aiPackage := runner.GenerateAIHelperPackage(result, cfg, insights, logTail, result.LogFile)
		_ = os.WriteFile(absHelpFilePath, []byte(aiPackage), 0644)

		// The crash card is the LAST thing printed to stdout so it remains directly in view
		report := runner.FormatDiagnosticReport(result, cfg, insights, absHelpFilePath)
		fmt.Println()
		fmt.Println(report)

		if result.AbortedByUser {
			os.Exit(130)
		}
		os.Exit(result.ExitCode)
	}

	fmt.Printf("\n[✓] Game session finished cleanly (duration: %v).\n", result.Duration.Round(100*time.Millisecond))
}
