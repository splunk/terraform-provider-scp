package utils

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"sort"
)

// IsSliceEqual function compares two lists ignoring the element order
func IsSliceEqual(in_a *[]string, in_b *[]string) bool {
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

// ParseSetValues converts the schema.Set object to an array of strings,
// it's supposed to be used for parsing resource data only
func ParseSetValues(values interface{}) []string {
	elementSet := values.(*schema.Set)
	parsedData := make([]string, 0)
	for _, elementValue := range elementSet.List() {
		parsedData = append(parsedData, elementValue.(string))
	}
	return parsedData
}
