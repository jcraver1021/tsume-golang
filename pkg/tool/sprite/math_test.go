package sprite

import (
	"testing"
)

func TestGCD(t *testing.T) {
	testCases := []struct {
		name string
		a, b int
		want int
	}{
		{"GCD of 0 and 0", 0, 0, 0},
		{"GCD of 0 and 5", 0, 5, 5},
		{"GCD of 5 and 0", 5, 0, 5},
		{"GCD of 6 and 9", 6, 9, 3},
		{"GCD of 20 and 30", 20, 30, 10},
		{"GCD of 17 and 13", 17, 13, 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := gcd(tc.a, tc.b); got != tc.want {
				t.Errorf("Expected GCD(%d, %d) = %d, got %d", tc.a, tc.b, tc.want, got)
			}
		})
	}
}

func TestLCM(t *testing.T) {
	testCases := []struct {
		name string
		a, b int
		want int
	}{
		{"LCM of 0 and 0", 0, 0, 0},
		{"LCM of 0 and 5", 0, 5, 0},
		{"LCM of 5 and 0", 5, 0, 0},
		{"LCM of 6 and 9", 6, 9, 18},
		{"LCM of 20 and 30", 20, 30, 60},
		{"LCM of 17 and 13", 17, 13, 221},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			if got := lcm(tc.a, tc.b); got != tc.want {
				t.Errorf("Expected LCM(%d, %d) = %d, got %d", tc.a, tc.b, tc.want, got)
			}
		})
	}
}
