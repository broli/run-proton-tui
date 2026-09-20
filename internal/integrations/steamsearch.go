package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"
)

type steamSearchResponse struct {
	Total int `json:"total"`
	Items []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	} `json:"items"`
}

// SearchSteamAppID queries the official Steam Store search endpoint using the game's folder name.
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
	req.Header.Set("User-Agent", "run-proton-tui/2.0")

	client := &http.Client{
		Timeout: 4 * time.Second,
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
