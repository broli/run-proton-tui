package integrations

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

// ProtonDBReport represents the compatibility summary retrieved from the ProtonDB API.
type ProtonDBReport struct {
	AppID            string `json:"appId"`
	Tier             string `json:"tier"`             // "platinum", "gold", "silver", "bronze", "borked"
	Total            int    `json:"total"`            // Total number of user reports
	TrendingTier     string `json:"trendingTier"`     // Recent reports tier
	BestReportedTier string `json:"bestReportedTier"` // Highest tier ever reported
	Confidence       string `json:"confidence"`       // "strong", "moderate", "weak"
}

// GetTierBadge returns a colored/styled text badge for display in the TUI.
func (r *ProtonDBReport) GetTierBadge() string {
	switch strings.ToLower(r.Tier) {
	case "platinum":
		return "⭐ PLATINUM"
	case "gold":
		return "🥇 GOLD"
	case "silver":
		return "🥈 SILVER"
	case "bronze":
		return "🥉 BRONZE"
	case "borked":
		return "❌ BORKED"
	default:
		return "❓ UNKNOWN"
	}
}

// FetchProtonDBReport queries ProtonDB's public summary REST API for a given AppID.
func FetchProtonDBReport(ctx context.Context, appID string) (*ProtonDBReport, error) {
	if appID == "" || appID == "0" {
		return nil, fmt.Errorf("invalid or missing AppID")
	}

	url := fmt.Sprintf("https://www.protondb.com/api/v1/reports/summaries/%s.json", appID)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", "run-proton-tui/2.0")

	client := &http.Client{
		Timeout: 4 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("no reports found for AppID %s", appID)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ProtonDB API returned status %d", resp.StatusCode)
	}

	var report ProtonDBReport
	if err := json.NewDecoder(resp.Body).Decode(&report); err != nil {
		return nil, err
	}

	return &report, nil
}
