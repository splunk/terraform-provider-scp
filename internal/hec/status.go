package hec

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sort"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

type HecBody struct {
	HttpEventCollector *v2.HecInfo `json:"http-event-collector"`
}

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
		return status.ProcessResponse(resp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
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
				ExpectedState: wait.TargetStatusResourceExists,
				LastError:     errors.New(string(bodyBytes)),
			}
		}

		var hec HecBody
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &hec); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		status := http.StatusText(resp.StatusCode)
		var hecSpec v2.HecSpec
		hecSpec = *hec.HttpEventCollector.Spec
		hecSpec.Token = hec.HttpEventCollector.Token
		return &hecSpec, status, nil
	}
}

// HecStatusDelete returns StateRefreshFunc that makes DELETE request and checks if request was accepted
func HecStatusDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DeleteHec(ctx, stack, v2.Hec(hecName), v2.DeleteHecJSONRequestBody{})
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
	}
}

// HecStatusUpdate returns StateRefreshFunc that makes PATCH request and checks if request was accepted
func HecStatusUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchHecRequest v2.PatchHECJSONRequestBody, hecName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		resp, err := acsClient.PatchHEC(ctx, stack, v2.Hec(hecName), patchHecRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, wait.TargetStatusResourceChange, wait.PendingStatusCRUD)
	}
}

// HecStatusVerifyUpdate returns a StateRefreshFunc that makes a GET request and checks to see if the hec fields matches those in patch request
func HecStatusVerifyUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchHECJSONRequestBody, hecName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeHec(ctx, stack, v2.Hec(hecName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != 200 {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{LastError: errors.New(string(bodyBytes))}
		}

		var hec HecBody
		var hecSpec v2.HecSpec
		updateComplete := false
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &hec); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
			hecSpec = *hec.HttpEventCollector.Spec //todo nil check
			updateComplete = VerifyHecUpdate(patchRequest, hecSpec)
		}

		var statusText string
		if updateComplete {
			statusText = status.UpdatedStatus
			return &hecSpec, statusText, nil
		} else {
			statusText = http.StatusText(resp.StatusCode)
			return nil, statusText, nil
		}
	}
}

// VerifyHecUpdate is a helper to verify that the fields in patch request match fields in the hec response
func VerifyHecUpdate(patchRequest v2.PatchHECJSONRequestBody, hec v2.HecSpec) bool {
	if patchRequest.AllowedIndexes != nil && !IsSpliceEqual(patchRequest.AllowedIndexes, hec.AllowedIndexes) {
		return false
	}
	if patchRequest.DefaultIndex != nil && (hec.DefaultIndex == nil || *patchRequest.DefaultIndex != *hec.DefaultIndex) {
		return false
	}
	if patchRequest.DefaultSource != nil && (hec.DefaultSource == nil || *patchRequest.DefaultSource != *hec.DefaultSource) {
		return false
	}
	if patchRequest.DefaultSourcetype != nil && (hec.DefaultSourcetype == nil || *patchRequest.DefaultSourcetype != *hec.DefaultSourcetype) {
		return false
	}
	if patchRequest.Disabled != nil && (hec.Disabled == nil || *patchRequest.Disabled != *hec.Disabled) {
		return false
	}

	if patchRequest.UseAck != nil && (hec.UseAck == nil || *patchRequest.UseAck != *hec.UseAck) {
		return false
	}
	return true
}

func IsSpliceEqual(in_a *[]string, in_b *[]string) bool {
	var a, b []string
	if in_a != nil {
		a = *in_a
	}
	if in_b != nil {
		b = *in_b
	}

	if len(a) != len(b) {
		return false
	}

	//Sort a and b to allow different ordering
	a = sort.StringSlice(a)
	b = sort.StringSlice(b)

	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
