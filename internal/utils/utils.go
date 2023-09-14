package utils

import "sort"

func IsSpliceEqual(in_a *[]string, in_b *[]string) bool {
	var a, b []string
	if in_a != nil {
		a = *in_a
	}
	if in_b != nil {
		b = *in_b
	}

	if len(a) != len(b) {
		return false
	}

	//Sort a and b to allow different ordering
	sort.Strings(a)
	sort.Strings(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
