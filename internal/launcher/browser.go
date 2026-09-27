package launcher

import (
	"fmt"
	"net/url"
	"os/exec"
	"strings"

	"github.com/atotto/clipboard"
)

// OpenURL launches the specified web address in the user's default desktop browser via xdg-open.
// It verifies that the URL uses an HTTP/HTTPS scheme and runs asynchronously without blocking the TUI.
func OpenURL(targetURL string) error {
	parsed, err := url.Parse(targetURL)
	if err != nil {
		return fmt.Errorf("invalid URL: %w", err)
	}
	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("refusing to open non-http/https URL: %s", targetURL)
	}

	xdgPath, err := exec.LookPath("xdg-open")
	if err != nil {
		return fmt.Errorf("xdg-open not found in system PATH (required for opening browser)")
	}

	cmd := exec.Command(xdgPath, targetURL)
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to launch browser via xdg-open: %w", err)
	}

	// Reap child process in the background to avoid leaking zombie (<defunct>) processes
	go func() {
		_ = cmd.Wait()
	}()

	return nil
}

// CopyToClipboard copies the given text to the system clipboard using the standard desktop clipboard bridge.
func CopyToClipboard(text string) error {
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("cannot copy empty text to clipboard")
	}

	if err := clipboard.WriteAll(text); err != nil {
		return fmt.Errorf("failed to copy text to clipboard: %w", err)
	}

	return nil
}
