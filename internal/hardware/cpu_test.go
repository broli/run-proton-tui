package hardware

import "testing"

func TestFormatCoreRange(t *testing.T) {
	tests := []struct {
		name     string
		cores    []int
		expected string
	}{
		{"empty", []int{}, ""},
		{"single core", []int{3}, "3"},
		{"range 0 to 11", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11}, "0-11"},
		{"unordered", []int{5, 2, 8, 1}, "1-8"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatCoreRange(tt.cores)
			if result != tt.expected {
				t.Errorf("formatCoreRange(%v) = %q; want %q", tt.cores, result, tt.expected)
			}
		})
	}
}

func TestDetectHardware(t *testing.T) {
	topo, err := DetectCPUTopology()
	if err != nil {
		t.Fatalf("DetectCPUTopology() returned unexpected error: %v", err)
	}
	if topo.TotalThreads <= 0 {
		t.Errorf("Expected TotalThreads > 0, got %d", topo.TotalThreads)
	}

	gpu, err := DetectGPU()
	if err != nil {
		t.Fatalf("DetectGPU() returned unexpected error: %v", err)
	}
	if gpu == nil {
		t.Fatal("Expected non-nil GPUInfo")
	}
}
