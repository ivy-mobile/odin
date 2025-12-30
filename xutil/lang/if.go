package lang

// If returns trueValue if condition is true, otherwise falseValue.
// same as ternary operator in C/C++
func If[T any](condition bool, trueValue, falseValue T) T {
	if condition {
		return trueValue
	}
	return falseValue
}
