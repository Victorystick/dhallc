package lib

// Maps to Dhall's # operator.
func ListConcat[T any](a, b []T) []T {
	res := make([]T, len(a)+len(b))

	copy(res, a)
	copy(res[len(a):], b)

	return res
}

func ListLength[T any](a []T) uint {
	return uint(len(a))
}
