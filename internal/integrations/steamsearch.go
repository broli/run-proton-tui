package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type steamSearchResponse struct {
	Total int `json:"total"`
	Items []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
}

// CleanGameTitle strips repack tags (e.g. [DODI Repack], (GOG), [FitGirl]), version numbers,
// and delimiters so Steam store search accurately matches the official game name.
func CleanGameTitle(raw string) string {
	reBrackets := regexp.MustCompile(`\[.*?\]|\(.*?\)|\{.*?\}`)
	cleaned := reBrackets.ReplaceAllString(raw, "")
	cleaned = strings.ReplaceAll(cleaned, "_", " ")
	cleaned = strings.TrimSpace(cleaned)
	cleaned = strings.TrimRight(cleaned, "-: ")
	if cleaned == "" {
		return raw
	}
	return cleaned
}

// SearchSteamAppID queries the official Steam Store search endpoint using the game's title.
// Returns the best matching AppID and Steam game name.
func SearchSteamAppID(ctx context.Context, query string) (string, string, error) {
	if query == "" {
		return "", "", fmt.Errorf("empty query string")
	}

	endpoint := fmt.Sprintf("https://store.steampowered.com/api/storesearch/?term=%s&l=english&cc=US", url.QueryEscape(query))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36")

	client := &http.Client{
		Timeout: 7 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", fmt.Errorf("Steam search returned HTTP status %d", resp.StatusCode)
	}

	var searchResp steamSearchResponse
	if err := json.NewDecoder(resp.Body).Decode(&searchResp); err != nil {
		return "", "", err
	}

	if len(searchResp.Items) == 0 {
		return "", "", fmt.Errorf("no Steam titles found for query %q", query)
	}

	best := searchResp.Items[0]
	return strconv.Itoa(best.ID), best.Name, nil
}

// ResolveSteamAppID tries multiple query candidates in sequence:
// 1. Preset name (e.g. "Arknights: Endfield")
// 2. Folder title (cleaned of [DODI Repack], [FitGirl], etc.)
// 3. Subdirectory of targetExe if nested (e.g. "games/Arknights Endfield/Endfield.exe" -> "Arknights Endfield")
// 4. Executable base name without .exe (e.g. "Endfield.exe" -> "Endfield")
func ResolveSteamAppID(ctx context.Context, presetName, targetExe, folderTitle string) (string, string, error) {
	var candidates []string

	if p := CleanGameTitle(presetName); p != "" {
		candidates = append(candidates, p)
	}
	if f := CleanGameTitle(folderTitle); f != "" && f != CleanGameTitle(presetName) {
		candidates = append(candidates, f)
	}

	if targetExe != "" {
		cleanExe := filepath.Clean(targetExe)
		dir := filepath.Dir(cleanExe)
		for dir != "." && dir != "/" && dir != "" {
			base := filepath.Base(dir)
			if !strings.EqualFold(base, "games") && !strings.EqualFold(base, "bin") && !strings.EqualFold(base, "game") {
				if cleaned := CleanGameTitle(base); cleaned != "" {
					candidates = append(candidates, cleaned)
				}
				break
			}
			dir = filepath.Dir(dir)
		}

		baseExe := strings.TrimSuffix(filepath.Base(targetExe), filepath.Ext(targetExe))
		if !strings.EqualFold(baseExe, "launcher") && !strings.EqualFold(baseExe, "setup") && !strings.EqualFold(baseExe, "installer") {
			if cleaned := CleanGameTitle(baseExe); cleaned != "" {
				candidates = append(candidates, cleaned)
			}
		}
	}

	var lastErr error
	seen := make(map[string]bool)
	for _, c := range candidates {
		c = strings.TrimSpace(c)
		if c == "" || seen[strings.ToLower(c)] {
			continue
		}
		seen[strings.ToLower(c)] = true

		aid, name, err := SearchSteamAppID(ctx, c)
		if err == nil && aid != "" && aid != "0" {
			return aid, name, nil
		}
		lastErr = err
	}

	if lastErr != nil {
		return "", "", lastErr
	}
	return "", "", fmt.Errorf("no candidate search queries available")
}
