package errors

import (
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

func IsUnknownFeatureError(err error) bool {
	switch v := err.(type) {
	case *resource.UnexpectedStateError:
		return strings.Contains(v.LastError.Error(), "unknown access feature")
	default:
		return false
	}
}
