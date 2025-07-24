package utils_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"github.com/stretchr/testify/assert"
)

func Test_TestIsSliceEqual(t *testing.T) {
	assert := assert.New(t)

	cases := []struct {
		expectedResult bool
		first          *[]string
		second         *[]string
	}{
		// Test Case 0: Expected true for empty
		{
			true,
			&[]string{},
			&[]string{},
		},
		// Test Case 1: Expected true for same array
		{
			true,
			&[]string{"a", "b"},
			&[]string{"a", "b"},
		},
		// Test Case 2: Expected false for different length
		{
			false,
			&[]string{"a"},
			&[]string{"b", "b"},
		},
		// Test Case 3: Expected true for different order
		{
			true,
			&[]string{"a", "b"},
			&[]string{"b", "a"},
		},
		// Test Case 4: Expected false for different entries
		{
			false,
			&[]string{"a", "b"},
			&[]string{"a", "a"},
		},
	}
	for i, test := range cases {
		test := test // Capture
		t.Run(fmt.Sprintf("case %d", i), func(_ *testing.T) {
			result := utils.IsSliceEqual(test.first, test.second)
			assert.Equal(result, test.expectedResult)
		})
	}
}

func Test_TestParseSetValues(t *testing.T) {
	testData := []interface{}{"a", "b", "c", "d"}
	values := schema.NewSet(schema.HashString, testData)
	parsedSet := utils.ParseSetValues(values)
	assert.ElementsMatch(t, parsedSet, testData)
}
