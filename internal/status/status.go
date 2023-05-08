package status

import (
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"io"
	"net/http"
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

	if !IsStatusCodeExpected(statusCode, targetStateCodes, pendingStatusCodes) {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, statusText, &resource.UnexpectedStateError{
			State:         statusText,
			ExpectedState: targetStateCodes,
			LastError:     errors.New(string(bodyBytes))}
	}
	return resp, statusText, nil
}
