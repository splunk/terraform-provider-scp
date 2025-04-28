package utils

import (
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

// IsSliceEqual function compares two lists ignoring the element order
func IsSliceEqual(inA *[]string, inB *[]string) bool {
	var a, b []string
	if inA != nil {
		a = *inA
	}
	if inB != nil {
		b = *inB
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
