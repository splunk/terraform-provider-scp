package ipallowlists

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
)

type allowlistResponse struct {
	Subnets []string
}

func IPAllowlistStatusCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, feature v2.Feature, subnets []string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		createBody := v2.AddSubnetsJSONRequestBody{
			Subnets: &subnets,
		}
		resp, err := acsClient.AddSubnets(ctx, stack, feature, createBody)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		return ProcessResponse(resp, TargetStatusResourceChange, PendingStatusCRUD)
	}
}

func IPAllowlistStatusRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, feature string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeAllowlist(ctx, stack, v2.Feature(feature))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if resp.StatusCode != http.StatusOK {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{
				State:         http.StatusText(resp.StatusCode),
				ExpectedState: TargetStatusResourceExists,
				LastError:     errors.New(string(bodyBytes)),
			}
		}

		var subnetsResponse allowlistResponse
		if resp.StatusCode == http.StatusOK {
			if err = json.Unmarshal(bodyBytes, &subnetsResponse); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		status := http.StatusText(resp.StatusCode)
		return subnetsResponse.Subnets, status, nil
	}
}

func ProcessResponse(resp *http.Response, targetStateCodes []string, pendingStatusCodes []string) (interface{}, string, error) {
	if resp == nil {
		return nil, "", &resource.UnexpectedStateError{LastError: errors.New("nil response")}
	}
	statusCode := resp.StatusCode
	statusText := http.StatusText(statusCode)

	if !status.IsStatusCodeExpected(statusCode, targetStateCodes, pendingStatusCodes) {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, statusText, &resource.UnexpectedStateError{
			State:         statusText,
			ExpectedState: targetStateCodes,
			LastError:     errors.New(string(bodyBytes))}
	}
	return resp, statusText, nil
}
