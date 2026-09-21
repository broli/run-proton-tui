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

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/diagnostics"
	"github.com/broli/run-proton-tui/internal/prefix"
	"github.com/broli/run-proton-tui/internal/runner"
	"github.com/broli/run-proton-tui/internal/ui"
	"github.com/broli/run-proton-tui/internal/ui/style"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-isatty"
)

var (
	version = "2.0.0"
)

func main() {
	flagNow := flag.Bool("now", false, "Launch immediately with saved or auto-detected settings (skip TUI)")
	flagNoTUI := flag.Bool("no-tui", false, "Alias for --now")
	flagYes := flag.Bool("y", false, "Alias for --now")
	flagClean := flag.Bool("clean", false, "Clean / reset Wine prefix before launch (with automatic save backup)")
	flagCleanShort := flag.Bool("c", false, "Alias for --clean")
	flagLog := flag.Bool("log", false, "Enable verbose Proton & DXVK logging to .logs/")
	flagLogShort := flag.Bool("v", false, "Alias for --log")
	flagDiagnostics := flag.Bool("diagnostics", false, "Run pre-flight health & permissions check and exit")
	flagVersion := flag.Bool("version", false, "Show version information")

	// Toggles
	flagGamescope := flag.String("gamescope", "", "Force Gamescope on/off (true/false)")
	flagPCores := flag.String("pcores", "", "Force CPU P-Core pinning on/off (true/false)")
	flagXalia := flag.String("xalia", "", "Force Proton Xalia on/off (true/false)")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "rpt (Run Proton TUI) v%s — Modular Linux Game Launcher Helper\n\n", version)
		fmt.Fprintf(os.Stderr, "Usage:\n")
		fmt.Fprintf(os.Stderr, "  rpt [flags] [optional-game.exe] [-- extra-game-args...]\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  rpt                    # Open interactive TUI in current game directory\n")
		fmt.Fprintf(os.Stderr, "  rpt --now              # Quick launch with saved/default settings\n")
		fmt.Fprintf(os.Stderr, "  rpt Game.exe           # Set target executable and open TUI\n")
		fmt.Fprintf(os.Stderr, "  rpt --clean --now      # Safe prefix reset followed by instant launch\n")
	}

	flag.Parse()

	if *flagVersion {
		fmt.Printf("rpt v%s\n", version)
		os.Exit(0)
	}

	gameDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting current working directory: %v\n", err)
		os.Exit(1)
	}

	// 1. Load or migrate configuration
	cfg, err := config.LoadConfig(gameDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error loading configuration: %v\n", err)
		os.Exit(1)
	}

	// Positional arguments: check for custom executable
	args := flag.Args()
	if len(args) > 0 {
		candidate := args[0]
		if strings.HasSuffix(strings.ToLower(candidate), ".exe") {
			cfg.TargetExe = candidate
			args = args[1:]
		}
	}

	// Apply CLI overrides to configuration
	if *flagClean || *flagCleanShort {
		pfxDir := filepath.Join(gameDir, "proton-prefix")
		res, err := prefix.SafeCleanPrefix(pfxDir, cfg.ProtonPath, gameDir, cfg.TargetExe, filepath.Base(gameDir))
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
	if *flagDiagnostics {
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

	skipTUI := *flagNow || *flagNoTUI || *flagYes || !isatty.IsTerminal(os.Stdin.Fd())

	if !skipTUI {
		// Run Interactive TUI
		m, err := ui.NewModel(gameDir, cfg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to initialize TUI: %v\n", err)
			os.Exit(1)
		}

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

	// 3. Execute Game Runner
	fmt.Printf("\n🚀 Launching %s with %s...\n", cfg.TargetExe, filepath.Base(cfg.ProtonPath))
	if cfg.UseGamescope {
		fmt.Printf("   Gamescope: Active (%dx%d @ %dHz -> %s)\n",
			cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeRefresh, cfg.GamescopeOutput)
	}
	if cfg.UsePCores {
		fmt.Printf("   CPU Cores: P-Cores pinned (Threads %s)\n", cfg.PCoresMask)
	}
	fmt.Println()

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
		sig, ok := <-sigChan
		if ok && sig != nil {
			cancel()
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

	// 4. Progressive Fallback, Crash & Stalled Session Interception
	if result.AbortedByUser || result.CrashDetected {
		insights := runner.AnalyzeSessionLog(result.LogFile, cfg)
		report := runner.FormatDiagnosticReport(result, cfg, insights)
		fmt.Println()
		fmt.Println(report)

		// Print Agent Prompt helper if user wants to copy-paste it
		logTail := runner.ReadLogTail(result.LogFile, 15)
		agentPrompt := runner.GenerateAgentPrompt(result, cfg, insights, logTail)
		fmt.Println("\n" + lipgloss.NewStyle().Bold(true).Foreground(style.ColorPrimary).Render("─── 📋 AGENT COPY-PASTE BLOCK (Markdown) ──────────────────────────────────"))
		fmt.Println(agentPrompt)
		fmt.Println(lipgloss.NewStyle().Bold(true).Foreground(style.ColorPrimary).Render("───────────────────────────────────────────────────────────────────────────"))

		if result.AbortedByUser {
			os.Exit(130)
		}
		os.Exit(result.ExitCode)
	}

	fmt.Printf("\n[✓] Game session finished cleanly (duration: %v).\n", result.Duration.Round(100*1000*1000))
}
