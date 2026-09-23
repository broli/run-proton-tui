# Fish shell completions for rpt (Run Proton TUI)
# Supports: rpt, rptui, run-proton

for cmd in rpt rptui run-proton
    # Clear existing completions for this command
    complete -c $cmd -e

    # Execution flags
    complete -c $cmd -l now -d "Launch immediately with saved or auto-detected settings (skip TUI)"
    complete -c $cmd -l no-tui -d "Bypass TUI and launch immediately"
    complete -c $cmd -s y -d "Bypass TUI and launch immediately"

    # Maintenance flags
    complete -c $cmd -l clean -d "Clean / reset Wine prefix before launch (with automatic save backup)"
    complete -c $cmd -s c -d "Clean / reset Wine prefix before launch (with automatic save backup)"

    # Diagnostics flags
    complete -c $cmd -l diagnostics -d "Run pre-flight health & permissions check and exit"
    complete -c $cmd -l diag -d "Run pre-flight health & permissions check and exit"
    complete -c $cmd -s d -d "Run pre-flight health & permissions check and exit"

    # Logging flags
    complete -c $cmd -l log -d "Enable verbose Proton & DXVK logging to .logs/"
    complete -c $cmd -s v -d "Enable verbose Proton & DXVK logging to .logs/"

    # Sandbox & Hardware overrides
    complete -c $cmd -l gamescope -x -a "true false" -d "Force Gamescope sandboxing on/off"
    complete -c $cmd -l pcores -x -a "true false" -d "Force CPU P-Core pinning on/off"
    complete -c $cmd -l xalia -x -a "true false" -d "Force Proton Xalia UI bridge on/off"

    # Information flags
    complete -c $cmd -l version -d "Show rpt version information"
    complete -c $cmd -l help -s h -d "Show help and command usage"

    # Executable argument completion (.exe files in current directory & subdirectories)
    complete -c $cmd -a "(find . -maxdepth 3 -name '*.exe' 2>/dev/null | sed 's|^\./||')" -d "Windows Executable"
end
