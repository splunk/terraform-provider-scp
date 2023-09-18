package errors

import (
	"net/http"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
)

const (
	ResourceExistsErr       = "already exists, use a different name to create hec token or use terraform import to bring current hec under terraform management"
	AcsErrSuffix            = "Please refer https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ACSerrormessages for general troubleshooting tips."
	FailedDeploymentTaskErr = "previous deployment task has failed"
)

func IsUnknownFeatureError(err error) bool {
	switch v := err.(type) {
	case *resource.UnexpectedStateError:
		return strings.Contains(v.LastError.Error(), "unknown access feature")
	default:
		return false
	}
}

func IsFailedDeploymentTaskError(err error) bool {
	switch v := err.(type) {
	case *resource.UnexpectedStateError:
		return strings.Contains(v.LastError.Error(), FailedDeploymentTaskErr)
	default:
		return false
	}
}

func IsConflictError(err error) bool {
	switch v := err.(type) {
	case *resource.UnexpectedStateError:
		return v.State == http.StatusText(http.StatusConflict)
	default:
		return false
	}
}
