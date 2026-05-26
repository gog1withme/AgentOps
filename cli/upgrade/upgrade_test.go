package upgrade

import "testing"

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		a, b string
		want int
	}{
		{"1.0.1", "1.0.0", 1},
		{"1.0.0", "1.0.1", -1},
		{"1.0.0", "1.0.0", 0},
		{"2.0.0", "1.9.9", 1},
	}
	for _, tc := range tests {
		got := compareVersions(tc.a, tc.b)
		if got != tc.want {
			t.Fatalf("compareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}
