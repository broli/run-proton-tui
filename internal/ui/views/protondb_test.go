package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/integrations"
	tea "github.com/charmbracelet/bubbletea"
)

func TestProtonDBView(t *testing.T) {
	// 1. Test empty report view
	pv := NewProtonDBView(nil, "GRYPHLINK", "0", 100, 30)
	pv.Status = "Could not find Steam AppID for 'GRYPHLINK'"
	vOut := pv.View()
	if !strings.Contains(vOut, "Could not find Steam AppID") {
		t.Errorf("Expected status in empty view output")
	}

	// 2. Test search mode activation
	query, refresh, contribute, done, cmd := pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'s'}})
	if !pv.SearchMode {
		t.Errorf("Expected SearchMode to be true after pressing 's'")
	}
	if query != "" || refresh || contribute || done || cmd == nil {
		t.Errorf("Unexpected return values on search activation")
	}

	// 3. Test typing and submitting custom query
	pv.SearchInput.SetValue("Endfield")
	query, refresh, _, _, _ = pv.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if pv.SearchMode {
		t.Errorf("Expected SearchMode to be false after pressing Enter")
	}
	if query != "Endfield" || !refresh {
		t.Errorf("Expected query='Endfield' and refresh=true, got query=%q, refresh=%v", query, refresh)
	}

	// 3b. Test 1-click contribute hotkey 'c'
	_, _, contribute, _, _ = pv.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'c'}})
	if !contribute {
		t.Errorf("Expected contribute=true after pressing 'c'")
	}

	// 4. Test view with loaded report
	rep := &integrations.ProtonDBReport{
		AppID:            "4732690",
		Tier:             "platinum",
		Total:            11,
		TrendingTier:     "platinum",
		BestReportedTier: "platinum",
		Confidence:       "moderate",
	}
	pvWithReport := NewProtonDBView(rep, "Arknights: Endfield", "4732690", 100, 30)
	repOut := pvWithReport.View()
	if !strings.Contains(repOut, "PLATINUM") {
		t.Errorf("Expected PLATINUM badge in report output")
	}
	if !strings.Contains(repOut, "11 user reports") {
		t.Errorf("Expected total reports in output")
	}
}
