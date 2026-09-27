package launcher

import (
	"testing"
)

func TestOpenURL_Validation(t *testing.T) {
	tests := []struct {
		url     string
		wantErr bool
	}{
		{"", true},
		{"ftp://example.com", true},
		{"file:///etc/passwd", true},
		{"javascript:alert(1)", true},
		{"https://www.protondb.com/app/4732690", false},
	}

	for _, tt := range tests {
		err := OpenURL(tt.url)
		if tt.wantErr && err == nil {
			t.Errorf("OpenURL(%q) expected error, got nil", tt.url)
		}
		// If wantErr is false, in a headless environment without xdg-open it might fail with "not found" which is fine,
		// but it shouldn't fail with scheme error.
		if !tt.wantErr && err != nil && err.Error() == "refusing to open non-http/https URL: "+tt.url {
			t.Errorf("OpenURL(%q) failed scheme validation: %v", tt.url, err)
		}
	}
}

func TestCopyToClipboard_Empty(t *testing.T) {
	err := CopyToClipboard("")
	if err == nil {
		t.Errorf("CopyToClipboard(\"\") expected error, got nil")
	}
}
