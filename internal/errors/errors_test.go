package errors

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/stretchr/testify/assert"
)

func Test_IsUnknownErrorFeature(t *testing.T) {
	t.Run("is not a Unexpected State Error", func(t *testing.T) {
		err := fmt.Errorf("this is some random error")
		got := IsUnknownFeatureError(err)
		assert.False(t, got)
	})

	t.Run("is not an unknown access feature", func(t *testing.T) {
		err := &resource.UnexpectedStateError{
			LastError: fmt.Errorf("some unknown error"),
		}
		got := IsUnknownFeatureError(err)
		assert.False(t, got)
	})

	t.Run("is an unknown access feature", func(t *testing.T) {
		err := &resource.UnexpectedStateError{
			LastError: fmt.Errorf("unknown access feature: mock-feature. The ACS API supports the following IP allow list features"),
		}
		got := IsUnknownFeatureError(err)
		assert.True(t, got)
	})
}
