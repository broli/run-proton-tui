package views

import (
	"strings"
	"testing"

	"github.com/broli/run-proton-tui/internal/proton"
	tea "github.com/charmbracelet/bubbletea"
)

func sampleRunners() []proton.Runner {
	return []proton.Runner{
		{
			Name:           "Proton-GE Latest",
			Path:           "/steam/compatibilitytools.d/Proton-GE Latest/proton",
			Type:           proton.RunnerTypeGE,
			IsGE:           true,
			Recommendation: "GE-Proton: High Compatibility",
		},
		{
			Name:           "GE-Proton11-7",
			Path:           "/steam/compatibilitytools.d/GE-Proton11-7/proton",
			Type:           proton.RunnerTypeGE,
			IsGE:           true,
			Recommendation: "GE-Proton: High Compatibility",
		},
		{
			Name:           "Proton-CachyOS Latest",
			Path:           "/steam/compatibilitytools.d/Proton-CachyOS Latest/proton",
			Type:           proton.RunnerTypeCachyOS,
			IsCachy:        true,
			Recommendation: "CachyOS: x86-64-v3/v4 Tuned",
		},
		{
			Name:           "dwproton-11.0-14",
			Path:           "/steam/compatibilitytools.d/dwproton-11.0-14/proton",
			Type:           proton.RunnerTypeDW,
			IsDW:           true,
			Recommendation: "Dawn Winery: Optimized for Anime Gacha",
		},
		{
			Name:           "Proton 10.0",
			Path:           "/steam/steamapps/common/Proton 10.0/proton",
			Type:           proton.RunnerTypeValve,
			IsValve:        true,
			Recommendation: "Valve Official: Stable Compatibility",
		},
		{
			Name:           "Proton - Experimental",
			Path:           "/steam/steamapps/common/Proton - Experimental/proton",
			Type:           proton.RunnerTypeExperimental,
			IsExperimental: true,
			IsValve:        true,
			Recommendation: "Valve Experimental: Bleeding Edge",
		},
	}
}

func TestProtonPickerNavigation(t *testing.T) {
	runners := sampleRunners()
	picker := NewProtonPickerView(runners, runners[1].Path)

	if picker.Cursor != 1 {
		t.Fatalf("Expected initial cursor to be 1 (for GE-Proton11-7), got %d", picker.Cursor)
	}

	// Move down
	picker, sel, cancel := picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if sel || cancel || picker.Cursor != 2 {
		t.Errorf("Expected cursor=2 after 'j', got cursor=%d, sel=%v, cancel=%v", picker.Cursor, sel, cancel)
	}

	// Move up
	picker, sel, cancel = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if sel || cancel || picker.Cursor != 1 {
		t.Errorf("Expected cursor=1 after 'k', got cursor=%d, sel=%v, cancel=%v", picker.Cursor, sel, cancel)
	}

	// Jump to bottom (End / G)
	picker, _, _ = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if picker.Cursor != len(runners)-1 {
		t.Errorf("Expected cursor=%d after 'G', got %d", len(runners)-1, picker.Cursor)
	}

	// Move down at bottom -> wrap to top
	picker, _, _ = picker.Update(tea.KeyMsg{Type: tea.KeyDown})
	if picker.Cursor != 0 {
		t.Errorf("Expected cursor=0 after wrap-down, got %d", picker.Cursor)
	}

	// Move up at top -> wrap to bottom
	picker, _, _ = picker.Update(tea.KeyMsg{Type: tea.KeyUp})
	if picker.Cursor != len(runners)-1 {
		t.Errorf("Expected cursor=%d after wrap-up, got %d", len(runners)-1, picker.Cursor)
	}

	// Jump to top (Home / g)
	picker, _, _ = picker.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if picker.Cursor != 0 {
		t.Errorf("Expected cursor=0 after 'g', got %d", picker.Cursor)
	}

	// Select with Enter
	picker, sel, cancel = picker.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if !sel || cancel {
		t.Fatalf("Expected sel=true, cancel=false after Enter")
	}
	if picker.Selected.Name != "Proton-GE Latest" {
		t.Errorf("Expected selected runner 'Proton-GE Latest', got %s", picker.Selected.Name)
	}

	// Cancel with Esc
	picker, sel, cancel = picker.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if sel || !cancel {
		t.Fatalf("Expected sel=false, cancel=true after Esc")
	}
}

func TestProtonPickerRendering(t *testing.T) {
	runners := sampleRunners()
	picker := NewProtonPickerView(runners, runners[3].Path) // dwproton-11.0-14 active
	picker.SetDimensions(80, 24)

	output := picker.View()

	// Check Title
	if !strings.Contains(output, "Select Proton Compatibility Tool") {
		t.Errorf("View missing title")
	}

	// Check Badges
	if !strings.Contains(output, "GE-Proton") {
		t.Errorf("View missing GE-Proton badge")
	}
	if !strings.Contains(output, "CachyOS") {
		t.Errorf("View missing CachyOS badge")
	}
	if !strings.Contains(output, "DW-Proton") {
		t.Errorf("View missing DW-Proton badge")
	}
	if !strings.Contains(output, "Valve Proton") {
		t.Errorf("View missing Valve Proton badge")
	}
	if !strings.Contains(output, "Experimental") {
		t.Errorf("View missing Experimental badge")
	}

	// Check Active Indicator
	if !strings.Contains(output, "✓ Active") {
		t.Errorf("View missing active indicator '✓ Active'")
	}
	if !strings.Contains(output, "[Currently Configured]") {
		t.Errorf("View preview card missing '[Currently Configured]'")
	}

	// Check preview footer contents
	if !strings.Contains(output, "Dawn Winery: Optimized for Anime Gacha") {
		t.Errorf("View missing recommendation in preview box")
	}
	if !strings.Contains(output, "Path:") {
		t.Errorf("View missing 'Path:' in preview box")
	}
}

func TestProtonPickerEmpty(t *testing.T) {
	picker := NewProtonPickerView([]proton.Runner{}, "")
	output := picker.View()
	if !strings.Contains(output, "No Proton runners found") {
		t.Errorf("Expected empty runners warning message, got: %s", output)
	}
}
