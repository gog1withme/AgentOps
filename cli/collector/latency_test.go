package collector

import "testing"

func TestComputeP95(t *testing.T) {
	tests := []struct {
		name string
		in   []int
		want float64
	}{
		{"empty", nil, 0},
		{"single", []int{100}, 100},
		{"twenty samples", []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20}, 19},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := computeP95(tc.in)
			if got != tc.want {
				t.Fatalf("computeP95(%v) = %v, want %v", tc.in, got, tc.want)
			}
		})
	}
}

func TestAppendLatencyWindow(t *testing.T) {
	var window []int
	for i := 1; i <= 105; i++ {
		window = appendLatencyWindow(window, i)
	}
	if len(window) != mcpLatencyWindowSize {
		t.Fatalf("expected window size %d, got %d", mcpLatencyWindowSize, len(window))
	}
	if window[0] != 6 || window[len(window)-1] != 105 {
		t.Fatalf("unexpected window bounds: first=%d last=%d", window[0], window[len(window)-1])
	}
}
