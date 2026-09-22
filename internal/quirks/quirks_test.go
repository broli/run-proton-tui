package quirks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestDetectQuirksCurated(t *testing.T) {
	// Endfield test
	preset := DetectQuirks("/home/carlos/Games/GRYPHLINK", "Endfield.exe", "0", "")
	if preset == nil {
		t.Fatalf("Expected Endfield preset to be detected")
	}
	if preset.UmuID != "umu-endfield" {
		t.Errorf("Expected UmuID=umu-endfield, got %s", preset.UmuID)
	}
	if preset.EnvVars["WINE_CANONICAL_HOLE"] != "skip_volatile_check" {
		t.Errorf("Expected WINE_CANONICAL_HOLE override")
	}
	if len(preset.Profiles) == 0 {
		t.Errorf("Expected dual profiles for Endfield")
	}
	if prof, ok := preset.Profiles["Launcher.exe"]; !ok || *prof.UseGamescope {
		t.Errorf("Expected Launcher.exe profile with UseGamescope=false")
	}

	// Repack installer test
	repack := DetectQuirks("/home/carlos/Games/SomeRepack", "Setup.exe", "0", "")
	if repack == nil {
		t.Fatalf("Expected Repack Installer preset to be detected")
	}
	if prof, ok := repack.Profiles["Setup.exe"]; !ok || *prof.UseGamescope {
		t.Errorf("Expected Setup.exe profile with UseGamescope=false")
	}
}

func TestScanLocalUMUDatabase(t *testing.T) {
	tmpDir := t.TempDir()
	pfixes := filepath.Join(tmpDir, "protonfixes")
	_ = os.MkdirAll(pfixes, 0755)

	csvContent := `TITLE,STORE,CODENAME,UMU_ID,COMMON ACRONYM (Optional),NOTE (Optional),EXE_STRINGS (Optional)
Arknights: Endfield,none,none,umu-endfield,,Standalone PC installer,
Dark Earth,none,none,umu-darkearth,,Standalone installer,darkearth.exe
`
	csvPath := filepath.Join(pfixes, "umu-database.csv")
	_ = os.WriteFile(csvPath, []byte(csvContent), 0644)

	mockProton := filepath.Join(tmpDir, "proton")

	preset := scanLocalUMUDatabase("dark_earth", "darkearth.exe", "0", mockProton)
	if preset == nil {
		t.Fatalf("Expected UMU CSV match for Dark Earth")
	}
	if preset.UmuID != "umu-darkearth" {
		t.Errorf("Expected umu-darkearth, got %s", preset.UmuID)
	}
}
