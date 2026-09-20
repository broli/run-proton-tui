package config

import (
	"bufio"
	"os"
	"strconv"
	"strings"
)

// ParseLegacyFishConfig reads an old Fish shell .proton-config key=value file
// and converts it into a modern GameConfig.
func ParseLegacyFishConfig(filePath string) (*GameConfig, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	cfg := NewDefaultConfig()
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if strings.HasPrefix(line, "#") || line == "" {
			continue
		}

		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])

		switch key {
		case "target_exe":
			cfg.TargetExe = val
		case "proton_path":
			cfg.ProtonPath = val
		case "use_gamescope":
			cfg.UseGamescope = (val == "1" || strings.ToLower(val) == "true")
		case "use_pcores":
			cfg.UsePCores = (val == "1" || strings.ToLower(val) == "true")
		case "manage_power":
			cfg.ManagePower = (val == "1" || strings.ToLower(val) == "true")
		case "use_xalia":
			cfg.UseXalia = (val == "1" || strings.ToLower(val) == "true")
		case "logging_enabled":
			cfg.EnableLogging = (val == "1" || strings.ToLower(val) == "true")
		case "gamescope_width":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.GamescopeWidth = n
			}
		case "gamescope_height":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.GamescopeHeight = n
			}
		case "gamescope_refresh":
			if n, err := strconv.Atoi(val); err == nil {
				cfg.GamescopeRefresh = n
			}
		case "gamescope_output":
			cfg.GamescopeOutput = val
		case "app_id":
			cfg.AppID = val
		}
	}

	return cfg, nil
}
