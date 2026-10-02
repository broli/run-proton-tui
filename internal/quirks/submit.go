package quirks

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/broli/run-proton-tui/internal/config"
	"github.com/broli/run-proton-tui/internal/launcher"
	"github.com/broli/run-proton-tui/internal/pathutil"
)

// MITConsentNotice is the mandatory licensing notice displayed to contributors.
const MITConsentNotice = `By submitting this quirk, launch configuration, or lifecycle hook to run-proton-tui, you certify that this is your own work and agree to release it under the MIT License.

Contributions will be curated by run-proton-tui maintainers and may be incorporated directly into the software, published openly, and submitted upstream to Valve's Proton, WineHQ, DXVK, VKD3D, or ProtonDB to improve Linux gaming compatibility for everyone.`

// SubmissionMethod identifies which mechanism was used to submit the quirk.
type SubmissionMethod string

const (
	MethodGH  SubmissionMethod = "gh"
	MethodGit SubmissionMethod = "git"
	MethodWeb SubmissionMethod = "web"
)

// SubmissionResult provides outcome details of the quirk submission attempt.
type SubmissionResult struct {
	Method  SubmissionMethod
	Message string
	URL     string
}

// GenerateQuirkRecipe formats the active game configuration and any lifecycle hooks into a clean Markdown submission.
func GenerateQuirkRecipe(gameDir string, cfg *config.GameConfig) (string, string) {
	gameTitle := filepath.Base(gameDir)
	slug := launcher.Slugify(gameTitle)

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("## Quirk Profile: %s\n\n", gameTitle))
	sb.WriteString(fmt.Sprintf("- **Game Title**: %s\n", gameTitle))
	sb.WriteString(fmt.Sprintf("- **Steam AppID**: `%s`\n", cfg.AppID))
	sb.WriteString(fmt.Sprintf("- **Target Executable**: `%s`\n", filepath.Base(cfg.TargetExe)))
	sb.WriteString(fmt.Sprintf("- **Recommended Proton**: `%s`\n", filepath.Base(cfg.ProtonPath)))
	sb.WriteString(fmt.Sprintf("- **Gamescope Recommended**: `%v`\n", cfg.UseGamescope))
	sb.WriteString(fmt.Sprintf("- **Prime-Run (NVIDIA)**: `%v`\n", cfg.UsePrimeRun))
	sb.WriteString(fmt.Sprintf("- **CPU P-Core Pinning**: `%v` (Mask: `%s`)\n", cfg.UsePCores, cfg.PCoresMask))

	if len(cfg.DLLOverrides) > 0 {
		sb.WriteString("\n### DLL Overrides\n```toml\n[dll_overrides]\n")
		for k, v := range cfg.DLLOverrides {
			sb.WriteString(fmt.Sprintf("%s = %q\n", k, v))
		}
		sb.WriteString("```\n")
	}

	if len(cfg.EnvVars) > 0 {
		sb.WriteString("\n### Environment Variables\n```bash\n")
		for k, v := range cfg.EnvVars {
			sb.WriteString(fmt.Sprintf("export %s=%q\n", k, v))
		}
		sb.WriteString("```\n")
	}

	// Check if a pre_launch hook exists
	hookPath := filepath.Join(gameDir, "hooks", "pre_launch.sh")
	if data, err := os.ReadFile(hookPath); err == nil && len(data) > 0 {
		sb.WriteString("\n### Lifecycle Hook (`hooks/pre_launch.sh`)\n```bash\n")
		sb.WriteString(pathutil.Sanitize(string(data)))
		sb.WriteString("\n```\n")
	}

	sb.WriteString("\n### Licensing Agreement\n")
	sb.WriteString("✓ I accept the MIT License terms and agree this contribution can be curated and submitted upstream to Valve's Proton / WineHQ.\n")

	return sb.String(), slug
}

// SubmitQuirk submits the current game quirk to broli/run-proton-tui adhering to the fallback order: gh -> git -> web.
func SubmitQuirk(gameDir string, cfg *config.GameConfig, acceptedMIT bool) (*SubmissionResult, error) {
	if !acceptedMIT {
		return nil, fmt.Errorf("submission rejected: contributor must accept the MIT License terms")
	}

	recipe, slug := GenerateQuirkRecipe(gameDir, cfg)
	title := fmt.Sprintf("feat(quirks): add curated profile for %s", filepath.Base(gameDir))

	// 1. Attempt GH CLI
	if ghPath, err := exec.LookPath("gh"); err == nil {
		authCmd := exec.Command(ghPath, "auth", "status")
		if err := authCmd.Run(); err == nil {
			branch := fmt.Sprintf("quirk/%s-%d", slug, time.Now().Unix())
			_ = exec.Command("git", "-C", gameDir, "checkout", "-b", branch).Run()

			// Create PR using gh CLI
			prCmd := exec.Command(ghPath, "pr", "create",
				"--repo", "broli/run-proton-tui",
				"--title", title,
				"--body", recipe,
			)
			prCmd.Dir = gameDir
			output, prErr := prCmd.CombinedOutput()
			if prErr == nil {
				prURL := strings.TrimSpace(string(output))
				return &SubmissionResult{
					Method:  MethodGH,
					Message: "Successfully submitted Pull Request via GitHub CLI (gh)!",
					URL:     prURL,
				}, nil
			}
		}
	}

	// 2. Attempt Local Git CLI (if inside a git workspace for run-proton-tui)
	if gitPath, err := exec.LookPath("git"); err == nil {
		isGitCmd := exec.Command(gitPath, "-C", gameDir, "rev-parse", "--is-inside-work-tree")
		if err := isGitCmd.Run(); err == nil {
			remoteOut, _ := exec.Command(gitPath, "-C", gameDir, "remote", "get-url", "origin").Output()
			if strings.Contains(string(remoteOut), "run-proton-tui") {
				branch := fmt.Sprintf("quirk/%s", slug)
				_ = exec.Command(gitPath, "-C", gameDir, "checkout", "-b", branch).Run()
				return &SubmissionResult{
					Method:  MethodGit,
					Message: fmt.Sprintf("Created local branch '%s'. Push your branch and open a PR to broli/run-proton-tui.", branch),
				}, nil
			}
		}
	}

	// 3. Universal Web Fallback (xdg-open + clipboard)
	_ = launcher.CopyToClipboard(recipe)

	baseURL := "https://github.com/broli/run-proton-tui/issues/new"
	params := url.Values{}
	params.Set("title", title)
	params.Set("body", recipe)
	targetURL := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	if err := launcher.OpenURL(targetURL); err != nil {
		return &SubmissionResult{
			Method:  MethodWeb,
			Message: "Quirk recipe copied to clipboard! Please open GitHub to submit an issue.",
			URL:     baseURL,
		}, nil
	}

	return &SubmissionResult{
		Method:  MethodWeb,
		Message: "✓ Quirk recipe copied to clipboard & opened GitHub in browser!",
		URL:     targetURL,
	}, nil
}
