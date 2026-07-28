package sprite

// gcd returns the greatest common divisor of a and b.
func gcd(a, b int) int {
	for b != 0 {
		a, b = b, a%b
	}
	return a
}

// lcm returns the least common multiple of a and b.
func lcm(a, b int) int {
	if a == 0 || b == 0 {
		return 0
	}
	return a * b / gcd(a, b)
}
