package status

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"io"
	"net/http"
	"strings"
)

const (
	UpdatedStatus           = "UPDATED"
	ErrIndexNotFound        = "404-index-not-found"
	ErrHecNotFound          = "404-hec-not-found"
	ErrUserNotFound         = "404-user-not-found"
	ErrRoleNotFound         = "404-role-not-found"
	ErrDependencyIncomplete = "424-dependency-incomplete"
	ErrFailedDependency     = "424-failed-dependency"
)

// IsStatusCodeExpected checks if the given status code exists in either target or pending status codes
func IsStatusCodeExpected(statusCode int, targetStatusCodes []string, pendingStatusCodes []string) bool {
	isRetryableError := false
	isTargetStatus := false

	for _, code := range targetStatusCodes {
		if code == http.StatusText(statusCode) {
			isTargetStatus = true
		}
	}

	for _, code := range pendingStatusCodes {
		if code == http.StatusText(statusCode) {
			isRetryableError = true
		}
	}

	return isTargetStatus || isRetryableError
}

func ProcessResponse(resp *http.Response, targetStateCodes []string, pendingStatusCodes []string) (interface{}, string, error) {
	if resp == nil {
		return nil, "", &resource.UnexpectedStateError{LastError: errors.New("nil response")}
	}
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)
	bodyBytes, _ := io.ReadAll(resp.Body)

	if !IsStatusCodeExpected(statusCode, targetStateCodes, pendingStatusCodes) {
		return nil, statusText, &resource.UnexpectedStateError{
			State:         statusText,
			ExpectedState: targetStateCodes,
			LastError:     errors.New(string(bodyBytes))}
	}

	// We will add logic to handle retry failed tasks or any general actions on behalf of the user here.
	if statusCode == http.StatusFailedDependency && !strings.Contains(string(bodyBytes), ErrDependencyIncomplete) {

		// We will catch and return any 424 error that is not a 424-dependency-incomplete and set more specific state
		return nil, statusText, &resource.UnexpectedStateError{
			State:         ErrDependencyIncomplete,
			ExpectedState: targetStateCodes,
			LastError:     errors.New(string(bodyBytes))}
	}
	return resp, statusText, nil

}
