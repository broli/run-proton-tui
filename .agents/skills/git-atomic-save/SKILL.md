---
name: git-atomic-save
description: >-
  Splits unstaged and staged changes into atomic, bite-sized Conventional Commits,
  runs pre-commit Go test verification, generates descriptive messages adhering to repository
  standards, and safely pushes to origin.
  Use when the user asks to "save to git", "saving to git", "save progress", "commit and push",
  "atomic commit", "commit my changes", "push changes", or types "/git-atomic-save" or "/save-to-git".
---

# Git Atomic Save Skill

This skill guides the agent in inspecting the working directory of `run-proton-tui` (`rpt`), verifying code quality with Go tests, organizing changes into coherent, bite-sized Conventional Commits, and safely pushing them to the remote branch on GitHub.

## Guiding Principles

1. **Atomic Slicing**: Every commit must represent a single logical change. Never bundle unrelated features, UI tweaks, lifecycle hooks, documentation updates, and build configs into a single catch-all commit.
2. **Quality Gate Before Commit**: Always run `go vet ./...` and `go test ./...` before committing. Never commit code that breaks the build or fails tests.
3. **Conventional Commits**: Strictly adhere to the Conventional Commits specification tailored for `run-proton-tui`:
   - `feat(<scope>): <imperative summary>` — New user-facing capability or system feature.
   - `fix(<scope>): <imperative summary>` — Bug fix or error resolution.
   - `refactor(<scope>): <imperative summary>` — Code restructuring without behavioral changes.
   - `test(<scope>): <imperative summary>` — Adding or updating test suites.
   - `docs(<scope>): <imperative summary>` — Documentation, wiki pages, README, or ROADMAP updates.
   - `chore(<scope>): <imperative summary>` — Dependencies (`go.mod`), Makefile, completions, build scripts.
4. **Valid Scopes for `run-proton-tui`**:
   - `runner`: Execution wrapper, exit tracking, crash detection (`internal/runner/`)
   - `ui`: Bubbletea models, dashboard, views, style theme (`internal/ui/`)
   - `hooks`: Lifecycle hook engine, cascade resolver, Chroma inspector (`internal/hooks/`)
   - `community-hooks`: Bundled game recipes and templates (`community-hooks/`)
   - `quirks`: Game presets, declarative TOML quirks engine (`internal/quirks/`)
   - `diagnostics`: Crash doctor, AI helper package, spec dump (`internal/diagnostics/`)
   - `launcher`: XDG `.desktop` application shortcut generator (`internal/launcher/`)
   - `prefix`: Isolated prefix management, clean/backup, registry overrides (`internal/prefix/`)
   - `config`: TOML configuration schemas, profiles, cloud sync (`internal/config/`)
   - `cli`: Main entrypoint flags, Linux boot status indicators (`cmd/rpt/`)
   - `hardware`: GPU detection, CPU topology, display outputs (`internal/hardware/`)
   - `integrations`: Steam emulators, ProtonDB integration (`internal/integrations/`)
5. **No Blind Staging**: Never run an unbounded `git add .` if changes belong to different logical concerns.

---

## Step-by-Step Procedure

### Step 1: Pre-flight Repository Inspection
1. Verify active branch:
   ```bash
   git branch --show-current
   ```
2. Inspect changed and untracked files:
   ```bash
   git status -s
   ```
3. If the working tree is already clean and up to date, notify the user that there is nothing to commit and stop.
4. Note: On this project, day-to-day work takes place on the `development` branch. If currently on `main`, inform the user that changes should typically be saved to `development` unless making an emergency hotfix.

### Step 2: Quality Gate & Test Verification
Run the project test suite and vet checks to ensure changes are stable:
```bash
go vet ./...
go test ./...
```
- If tests or vet fail, **HALT immediately**.
- Display the failing test/vet output and fix the issue before staging any commits.

### Step 3: Analyze Diffs and Group into Logical Units
1. Review unstaged and staged diffs:
   ```bash
   git diff
   git status -u
   ```
2. Cluster modified/untracked files into atomic units based on their scope:
   - **Runner & Crash Doctor**: `internal/runner/` ➔ `feat(runner)` or `fix(runner)`
   - **UI & Dashboard**: `internal/ui/` ➔ `feat(ui)` or `fix(ui)`
   - **Lifecycle Hooks & Community**: `internal/hooks/`, `community-hooks/` ➔ `feat(hooks)`
   - **Declarative Quirks**: `internal/quirks/` ➔ `feat(quirks)`
   - **Desktop Shortcuts**: `internal/launcher/` ➔ `feat(launcher)`
   - **Diagnostics & Spec Dump**: `internal/diagnostics/` ➔ `feat(diagnostics)`
   - **Configuration & Cloud Sync**: `internal/config/` ➔ `feat(config)`
   - **CLI & Flags**: `cmd/rpt/main.go` ➔ `feat(cli)`
   - **Documentation**: `README.md`, `ROADMAP.md`, `docs/*` ➔ `docs`
   - **Build & Tooling**: `Makefile`, `go.mod`, `go.sum`, completions ➔ `chore`

### Step 4: Stage and Commit Each Logical Unit
For each logical cluster identified:
1. Stage only the relevant files for this unit:
   ```bash
   git add <path/to/file1> <path/to/file2>
   ```
2. Compose a high-quality Conventional Commit message:
   - **Subject**: Concise imperative sentence (under 72 chars), e.g.:
     - `feat(hooks): add chroma syntax-highlighted hook inspector and pager`
     - `feat(launcher): add 1-click xdg desktop shortcut generator`
     - `fix(runner): unmask child process exit code from gamescope`
     - `docs: update roadmap and philosophy in readme`
   - **Body** (for non-trivial changes): Bullet points highlighting *what* was changed and *why*.
3. Execute the commit:
   ```bash
   git commit -m "<type>(<scope>): <summary>" -m "<detailed body bullets>"
   ```
4. Verify remaining files with `git status -s` and repeat for the next cluster until the working tree is completely clean.

### Step 5: Push to Current Working Branch
1. Push all newly created commits to origin:
   ```bash
   git push -u origin $(git branch --show-current)
   ```
2. Verify push status:
   ```bash
   git status
   ```

### Step 6: Summary Report
Print a clear summary table for the user:
- Total commits created and pushed
- Commit hashes with their Conventional Commit titles
- Working branch name (`development` or feature branch)
- Remote repository synchronization confirmation
