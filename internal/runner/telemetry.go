package runner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
)

// SessionRecord represents an anonymized, local-only snapshot of a single game session.
// It is stored strictly on the local filesystem and is NEVER transmitted over any network.
type SessionRecord struct {
	Timestamp       time.Time `json:"timestamp"`
	AppID           string    `json:"app_id"`
	GameTitle       string    `json:"game_title"`
	TargetExe       string    `json:"target_exe"`
	DurationSeconds float64   `json:"duration_seconds"`
	ExitCode        int       `json:"exit_code"`
	CrashDetected   bool      `json:"crash_detected"`
	AbortedByUser   bool      `json:"aborted_by_user"`
	RunnerName      string    `json:"runner_name"`
	QuirksApplied   []string  `json:"quirks_applied"`
	HardwareSummary string    `json:"hardware_summary"`
}

// SanitizePath scrubs absolute paths containing the user's home directory,
// replacing "/home/username/..." with "~/..." to guarantee total privacy.
func SanitizePath(p string) string {
	if p == "" {
		return ""
	}
	home, err := os.UserHomeDir()
	if err == nil && home != "" && strings.HasPrefix(p, home) {
		return "~" + strings.TrimPrefix(p, home)
	}
	return p
}

// GetHistoryFilePath returns the canonical local path to the per-game session history ledger.
func GetHistoryFilePath(gameDir string) string {
	return filepath.Join(gameDir, ".logs", "session-history.json")
}

// RecordSession appends a new session entry to the local JSON history ledger.
// CRITICAL PRIVACY RULE: If cfg.EnableLocalTelemetry is false, this function
// exits immediately and writes absolutely nothing to disk (Opt-In / Default OFF).
func RecordSession(gameDir string, cfg *config.GameConfig, result *SessionResult, hwSummary string) error {
	if cfg == nil || !cfg.EnableLocalTelemetry {
		// Strict default: zero persistent tracking when disabled
		return nil
	}

	if result == nil {
		return nil
	}

	historyFile := GetHistoryFilePath(gameDir)
	logsDir := filepath.Dir(historyFile)
	if err := os.MkdirAll(logsDir, 0755); err != nil {
		return fmt.Errorf("failed to create logs directory: %w", err)
	}

	history, err := ReadSessionHistory(gameDir)
	if err != nil {
		history = make([]SessionRecord, 0)
	}

	// Identify quirks applied
	quirks := make([]string, 0)
	if cfg.UseGamescope {
		quirks = append(quirks, fmt.Sprintf("gamescope (%s, %dx%d@%dHz)", cfg.GamescopeOutput, cfg.GamescopeWidth, cfg.GamescopeHeight, cfg.GamescopeRefresh))
	}
	if cfg.UsePCores {
		quirks = append(quirks, fmt.Sprintf("p-cores (%s)", cfg.PCoresMask))
	}
	if cfg.UsePrimeRun {
		quirks = append(quirks, "prime-run")
	}
	if cfg.UseXalia {
		quirks = append(quirks, "xalia")
	}
	if len(cfg.DLLOverrides) > 0 {
		quirks = append(quirks, fmt.Sprintf("%d dll overrides", len(cfg.DLLOverrides)))
	}
	if cfg.PresetName != "" {
		quirks = append(quirks, fmt.Sprintf("preset: %s", cfg.PresetName))
	}

	runnerName := filepath.Base(cfg.ProtonPath)
	if runnerName == "" || runnerName == "." {
		runnerName = "Default Proton"
	}

	record := SessionRecord{
		Timestamp:       time.Now().UTC(),
		AppID:           cfg.AppID,
		GameTitle:       filepath.Base(gameDir),
		TargetExe:       SanitizePath(cfg.TargetExe),
		DurationSeconds: result.Duration.Seconds(),
		ExitCode:        result.ExitCode,
		CrashDetected:   result.CrashDetected,
		AbortedByUser:   result.AbortedByUser,
		RunnerName:      runnerName,
		QuirksApplied:   quirks,
		HardwareSummary: hwSummary,
	}

	// Append record and cap at last 50 sessions
	history = append(history, record)
	if len(history) > 50 {
		history = history[len(history)-50:]
	}

	data, err := json.MarshalIndent(history, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to encode session history: %w", err)
	}

	if err := os.WriteFile(historyFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session history to %s: %w", historyFile, err)
	}

	return nil
}

// ReadSessionHistory loads the local session history ledger from disk.
// Returns an empty slice and nil error if no history file currently exists.
func ReadSessionHistory(gameDir string) ([]SessionRecord, error) {
	historyFile := GetHistoryFilePath(gameDir)
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			return make([]SessionRecord, 0), nil
		}
		return nil, err
	}

	var history []SessionRecord
	if err := json.Unmarshal(data, &history); err != nil {
		return nil, fmt.Errorf("corrupt session history JSON: %w", err)
	}

	return history, nil
}

// ClearSessionHistory deletes the local session history ledger file.
func ClearSessionHistory(gameDir string) error {
	historyFile := GetHistoryFilePath(gameDir)
	if err := os.Remove(historyFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session history: %w", err)
	}
	return nil
}
