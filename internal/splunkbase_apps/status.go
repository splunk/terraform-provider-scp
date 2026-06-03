package splunkbaseapps

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

// AppStatusCreate returns StateRefreshFunc that makes POST request and checks if response is accepted
func AppStatusCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, params v2.InstallAppVictoriaParams, body io.Reader) resource.StateRefreshFunc {

	return func() (interface{}, string, error) {
		contentType := "application/x-www-form-urlencoded"
		resp, err := acsClient.InstallAppVictoriaWithBody(ctx, stack, &params, contentType, body)

		if err != nil && (resp == nil || resp.StatusCode != http.StatusConflict) {
			return nil, responseStatus(resp), &resource.UnexpectedStateError{LastError: err}
		}
		if err := validateResponse(resp); err != nil {
			return nil, "", err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
			}
		}(resp.Body)
		return status.ProcessResponse(resp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
	}
}

var GeneralRetryableStatusCodes = map[int]string{
	http.StatusTooManyRequests:    http.StatusText(http.StatusTooManyRequests),
	http.StatusServiceUnavailable: http.StatusText(http.StatusServiceUnavailable),
}

func responseStatus(resp *http.Response) string {
	if resp == nil {
		return ""
	}

	return resp.Status
}

func validateResponse(resp *http.Response) error {
	if resp == nil {
		return &resource.UnexpectedStateError{LastError: errors.New("nil response")}
	}

	return nil
}

func AppStatusRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName v2.AppName) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.DescribeAppVictoria(ctx, stack, appName)
		if err != nil {
			return nil, responseStatus(resp), &resource.UnexpectedStateError{LastError: err}
		}
		if err := validateResponse(resp); err != nil {
			return nil, "", err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
			}
		}(resp.Body)
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != http.StatusOK {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{
				State:         http.StatusText(resp.StatusCode),
				ExpectedState: wait.TargetStatusResourceExists,
				LastError:     errors.New(string(bodyBytes)),
			}
		}

		var appResponse v2.DescribeAppVictoriaResponse
		if resp.StatusCode == http.StatusOK {
			if err = json.Unmarshal(bodyBytes, &appResponse.JSON200); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		statusText := http.StatusText(resp.StatusCode)
		return appResponse.JSON200, statusText, nil
	}
}

func AppStatusDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName v2.AppName, params v2.UninstallAppVictoriaParams) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.UninstallAppVictoria(ctx, stack, appName, &params)
		if err != nil {
			return nil, responseStatus(resp), &resource.UnexpectedStateError{LastError: err}
		}
		if err := validateResponse(resp); err != nil {
			return nil, "", err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
			}
		}(resp.Body)
		return status.ProcessResponse(resp, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
	}
}

func AppStatusUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName v2.AppName, params v2.PatchAppVictoriaParams, body io.Reader) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.PatchAppVictoriaWithBody(ctx, stack, appName, &params, "application/x-www-form-urlencoded", body)
		if err != nil {
			return nil, responseStatus(resp), &resource.UnexpectedStateError{LastError: err}
		}
		if err := validateResponse(resp); err != nil {
			return nil, "", err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
			}
		}(resp.Body)
		return status.ProcessResponse(resp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
	}
}

func AppStatusPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName v2.AppName, targetStatus []string, pendingStatus []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.DescribeAppVictoria(ctx, stack, appName)
		if err != nil {
			return nil, responseStatus(resp), &resource.UnexpectedStateError{LastError: err}
		}
		if err := validateResponse(resp); err != nil {
			return nil, "", err
		}
		defer func(Body io.ReadCloser) {
			err := Body.Close()
			if err != nil {
				tflog.Error(ctx, fmt.Sprintf("Error closing response body: %v", err))
			}
		}(resp.Body)

		// Retry on 429 or 503
		if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode == http.StatusServiceUnavailable {
			return nil, http.StatusText(resp.StatusCode), fmt.Errorf("retryable status code: %d", resp.StatusCode)
		}

		return status.ProcessResponse(resp, targetStatus, pendingStatus)
	}
}
