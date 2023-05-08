package hec

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"io"
	"net/http"
)

var GeneralRetryableStatusCodes = map[int]string{
	http.StatusTooManyRequests: http.StatusText(429),
}

// HecStatusCreate returns StateRefreshFunc that makes POST request and checks if response is accepted
func HecStatusCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createHecRequest v2.CreateHECJSONRequestBody) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.CreateHEC(ctx, stack, createHecRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		return status.ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

// HecStatusPoll returns StateRefreshFunc that makes GET request and checks if response is desired target (200 for create and 404 for delete)
func HecStatusPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string, targetStatus []string, pendingStatus []string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeHec(ctx, stack, v2.Hec(hecName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, targetStatus, pendingStatus)
	}
}

// HecStatusRead returns StateRefreshFunc that makes GET request, checks if request was successful, and returns hec response
func HecStatusRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeHec(ctx, stack, v2.Hec(hecName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != http.StatusOK {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{
				State:         http.StatusText(resp.StatusCode),
				ExpectedState: TargetStatusResourceExists,
				LastError:     errors.New(string(bodyBytes)),
			}
		}

		var hec v2.HecSpec
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &hec); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		status := http.StatusText(resp.StatusCode)
		return &hec, status, nil
	}
}
