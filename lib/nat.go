package lib

import "strconv"

func NaturalShow(n uint) string {
	return strconv.FormatUint(uint64(n), 10)
}

func NaturalSubtract(a uint) func(b uint) uint {
	return func(b uint) uint {
		if a < b {
			return 0
		}
		return a - b
	}
}

func NaturalToInteger(n uint) int {
	return int(n)
}
