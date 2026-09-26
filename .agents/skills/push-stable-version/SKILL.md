---
name: push-stable-version
description: >-
  Validates repository test suite, verifies clean build, creates a GitHub Pull Request from
  development into main using gh CLI, checks CI/mergeability, performs squash-merge, tags release,
  and updates documentation.
  Use when the user asks to "push new stable version", "pushing new stable version", "release stable version",
  "ship stable version", "ship new version", "merge to main", "release to main", or types "/push-stable-version" or "/release-stable-version".
---

# Push Stable Version Skill

This skill orchestrates the end-to-end release pipeline for `run-proton-tui` (`rpt`). It validates code quality across the entire Go test suite, builds the production binary, creates a GitHub Pull Request from `development` targeting the stable `main` branch via the `gh` CLI, performs a verified merge, tags the new release, and ensures local and remote branches are cleanly synchronized.

## Workflow Rules & Safety Rails

1. **Branch Architecture**:
   - **`development`**: The active integration branch where all new features, bug fixes, and tests land.
   - **`main`**: The **stable production release branch**. Direct pushes to `main` are restricted; releases merge through verified Pull Requests.
   - **Branch Preservation**: **NEVER delete the `development` branch** during or after merge. Unlike temporary feature branches, `development` is permanent.
2. **Zero-Tolerance Quality Gate**:
   - `go vet ./...` must pass with zero warnings.
   - `go test -count=1 ./...` must pass 100% without cached test results.
   - `make build` must produce a valid binary in `bin/rpt`.
   - Smoke tests must succeed (`./bin/rpt --version`, `./bin/rpt --help`, `./bin/rpt --inspect-hooks`).
3. **GitHub CLI (`gh`) Integration**:
   - All GitHub remote operations (PR creation, checks verification, merging, sync) must use the authenticated `gh` CLI.
4. **Clean Working Tree Prerequisite**:
   - The working tree on `development` must be clean before creating the release. If unstaged or staged changes exist, run the **`git-atomic-save`** skill first.

---

## Step-by-Step Procedure

### Step 1: Pre-flight Verification & Tree Cleanliness
1. Verify active branch is `development`:
   ```bash
   git branch --show-current
   ```
   If currently on another branch, check out `development`:
   ```bash
   git checkout development
   ```
2. Check working tree status:
   ```bash
   git status -s
   ```
3. If uncommitted changes exist:
   - Automatically execute the **`git-atomic-save`** workflow to group changes into Conventional Commits and push to `origin/development`.
4. Ensure `development` is fully pushed and up to date with remote:
   ```bash
   git push origin development
   ```

### Step 2: Strict Quality Gate & Build Verification
Execute the full test suite and build verification:
```bash
go vet ./...
go test -count=1 ./...
make build
make check-deps
```
Run binary smoke tests:
```bash
./bin/rpt --version
./bin/rpt --help
./bin/rpt --inspect-hooks
```
- If any test, vet check, or build step fails, **HALT immediately**.
- Report the failure details to the user and resolve all issues before proceeding.

### Step 3: Determine Version & Prepare Release Notes
1. Extract the current version string from `cmd/rpt/main.go`:
   ```bash
   grep 'version = ' cmd/rpt/main.go
   ```
2. Confirm the target version tag (e.g. `v0.5.0` or `v0.5.0-alpha`).
3. Summarize key achievements since the last release on `main`:
   ```bash
   git log --oneline main..development
   ```
4. Draft a structured release summary containing:
   - **🚀 New Capabilities**: Promoted executable switcher (`[e / 2]`), 1-click desktop shortcuts (`[s]`), Linux boot status lines.
   - **🩺 Crash Doctor & AI Diagnostics**: Gamescope child exit unmasking, false-positive-friendly 90s window, modular `signatures.toml`, `.logs/ask-ai-help.txt` generator with absolute paths.
   - **🪝 Lifecycle Hooks & Chroma Inspector**: Cascade discovery, `$RPT_*` environment cheatsheet, syntax-highlighted bash script pager, and community recipes.
   - **🧩 Declarative Quirks Engine**: TOML quirks manifests (`.rpt/quirks.toml`).
   - **☁️ Cloud Save Sync**: `[cloud_sync]` configuration with `rclone` / `syncthing`.
   - **🧪 Test Suite**: 100% test passing status across all packages.

### Step 4: Check Existing PR or Create via `gh`
1. Check if a PR from `development` into `main` already exists:
   ```bash
   gh pr list --base main --head development --json number,title,url,state
   ```
2. If no open PR exists, create one:
   ```bash
   gh pr create --base main --head development \
     --title "release: $(grep 'version = ' cmd/rpt/main.go | cut -d'"' -f2) stable release" \
     --body "<release_summary_body>"
   ```
3. Retrieve PR number and URL:
   ```bash
   gh pr view --json number,title,url,mergeable,mergeStateStatus
   ```

### Step 5: Verify PR Checks & Mergeability
1. Check CI status and mergeability:
   ```bash
   gh pr checks
   gh pr view --json mergeable,mergeStateStatus
   ```
2. If mergeable status is clean, proceed to merge.
3. If conflicts are reported, fetch `origin/main` and resolve conflicts on `development`, run tests, and push back to `origin/development`.

### Step 6: Squash-Merge into `main`
Merge the PR into `main` using `gh`:
```bash
# Note: Do NOT use --delete-branch, because 'development' must be preserved!
gh pr merge --squash --auto || gh pr merge --squash
```

### Step 7: Update Local `main` and Tag Release
1. Switch to `main` and pull the newly merged release:
   ```bash
   git checkout main
   git pull origin main
   ```
2. Create and push the annotated Git release tag:
   ```bash
   VERSION=$(grep 'version = ' cmd/rpt/main.go | cut -d'"' -f2)
   TAG="v${VERSION}"
   git tag -a "${TAG}" -m "Release ${TAG}: Modular Linux Game Launcher Helper"
   git push origin "${TAG}"
   ```

### Step 8: Return to `development` and Synchronize
1. Switch back to `development`:
   ```bash
   git checkout development
   ```
2. Merge or rebase with `main` so `development` is aligned with the release commit:
   ```bash
   git merge main -m "chore: sync development with release ${TAG} from main"
   git push origin development
   ```

### Step 9: Optional Wiki Documentation Sync
If documentation was updated, publish the wiki pages:
```bash
make publish-wiki || true
```

### Step 10: Release Completion Report
Print a release card summarizing:
- **Release Version & Tag**: e.g., `v0.5.0`
- **GitHub PR URL**: Link to merged PR
- **Target Branch**: `main` (Production)
- **Active Branch**: `development` (Ready for continued development)
- **Release Highlights**: Key features and fixes deployed
