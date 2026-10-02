package style

import "testing"

func TestClampWidth(t *testing.T) {
	tests := []struct {
		name       string
		totalWidth int
		margin     int
		minWidth   int
		maxWidth   int
		expected   int
	}{
		{
			name:       "within bounds",
			totalWidth: 90,
			margin:     10,
			minWidth:   60,
			maxWidth:   88,
			expected:   80,
		},
		{
			name:       "clamps below min",
			totalWidth: 50,
			margin:     10,
			minWidth:   60,
			maxWidth:   88,
			expected:   60,
		},
		{
			name:       "clamps above max",
			totalWidth: 120,
			margin:     10,
			minWidth:   60,
			maxWidth:   88,
			expected:   88,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := ClampWidth(tt.totalWidth, tt.margin, tt.minWidth, tt.maxWidth)
			if got != tt.expected {
				t.Errorf("ClampWidth(%d, %d, %d, %d) = %d, expected %d",
					tt.totalWidth, tt.margin, tt.minWidth, tt.maxWidth, got, tt.expected)
			}
		})
	}
}

func TestCapitalize(t *testing.T) {
	if got := Capitalize(""); got != "" {
		t.Errorf("Capitalize(\"\") = %q, expected \"\"", got)
	}
	if got := Capitalize("hello"); got != "Hello" {
		t.Errorf("Capitalize(\"hello\") = %q, expected \"Hello\"", got)
	}
}
