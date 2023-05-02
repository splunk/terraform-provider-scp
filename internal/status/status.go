package status

import "net/http"

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
