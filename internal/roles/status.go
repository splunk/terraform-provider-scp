package roles

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

var GeneralRetryableStatusCodes = map[int]string{
	http.StatusTooManyRequests: http.StatusText(http.StatusTooManyRequests),
}

// RoleStatusCreate returns StateRefreshFunc that makes POST request and checks if response is accepted
func RoleStatusCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createParams v2.CreateRoleParams, createRoleRequest v2.CreateRoleJSONRequestBody) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.CreateRole(ctx, stack, &createParams, createRoleRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		return status.ProcessResponse(resp, wait.TargetStatusResourceExists, wait.PendingStatusCRUD)
	}
}

// RoleStatusRead returns StateRefreshFunc that makes GET request, checks if request was successful, and returns role response
func RoleStatusRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, roleName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeRole(ctx, stack, v2.RoleName(roleName))
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

		var role v2.RolesResponse
		if resp.StatusCode == http.StatusOK {
			if err = json.Unmarshal(bodyBytes, &role); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		status := http.StatusText(resp.StatusCode)
		return &role, status, nil
	}
}

// RoleStatusDelete returns StateRefreshFunc that makes DELETE request and checks if request was accepted
func RoleStatusDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, roleName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DeleteRole(ctx, stack, v2.RoleName(roleName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, wait.TargetStatusResourceExists, wait.PendingStatusCRUD)
	}
}

// RoleStatusUpdate returns StateRefreshFunc that makes PATCH request and checks if request was successful
func RoleStatusUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchParams v2.PatchRoleInfoParams, patchRoleRequest v2.PatchRoleInfoJSONRequestBody, roleName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		resp, err := acsClient.PatchRoleInfo(ctx, stack, v2.RoleName(roleName), &patchParams, patchRoleRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, TargetStatusResourceChange, wait.PendingStatusCRUD)
	}
}

// RoleStatusVerifyUpdate returns a StateRefreshFunc that makes a GET request and checks to see if the role fields matches those in patch request
func RoleStatusVerifyUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchRoleInfoJSONRequestBody, roleName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeRole(ctx, stack, v2.RoleName(roleName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != http.StatusOK {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{LastError: errors.New(string(bodyBytes))}
		}

		var roleResponse v2.RolesResponse
		updateComplete := false
		if resp.StatusCode == http.StatusOK {
			if err = json.Unmarshal(bodyBytes, &roleResponse); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
			updateComplete = VerifyRoleUpdate(patchRequest, roleResponse)
		}

		var statusText string
		if updateComplete {
			statusText = status.UpdatedStatus
			return &roleResponse, statusText, nil
		}
		statusText = http.StatusText(resp.StatusCode)
		return nil, statusText, nil
	}
}

// VerifyRoleUpdate is a helper to verify that the fields in patch request match fields in the role response
func VerifyRoleUpdate(patchRequest v2.PatchRoleInfoJSONRequestBody, roleResponse v2.RolesResponse) bool {
	if patchRequest.CumulativeRTSrchJobsQuota != nil && (roleResponse.CumulativeRTSrchJobsQuota == nil || *patchRequest.CumulativeRTSrchJobsQuota != *roleResponse.CumulativeRTSrchJobsQuota) {
		return false
	}
	if patchRequest.CumulativeSrchJobsQuota != nil && (roleResponse.CumulativeSrchJobsQuota == nil || *patchRequest.CumulativeSrchJobsQuota != *roleResponse.CumulativeSrchJobsQuota) {
		return false
	}
	if patchRequest.DefaultApp != nil && (roleResponse.DefaultApp == nil || *patchRequest.DefaultApp != *roleResponse.DefaultApp) {
		return false
	}
	if patchRequest.ImportedRoles != nil && !utils.IsSliceEqual(patchRequest.ImportedRoles, roleResponse.Imported.Roles) {
		return false
	}
	if patchRequest.RtSrchJobsQuota != nil && (roleResponse.RtSrchJobsQuota == nil || *patchRequest.RtSrchJobsQuota != *roleResponse.RtSrchJobsQuota) {
		return false
	}
	if patchRequest.SrchJobsQuota != nil && (roleResponse.SrchJobsQuota == nil || *patchRequest.SrchJobsQuota != *roleResponse.SrchJobsQuota) {
		return false
	}
	if patchRequest.SrchDiskQuota != nil && (roleResponse.SrchDiskQuota == nil || *patchRequest.SrchDiskQuota != *roleResponse.SrchDiskQuota) {
		return false
	}
	if patchRequest.SrchFilter != nil && (roleResponse.SrchFilter == nil || *patchRequest.SrchFilter != *roleResponse.SrchFilter) {
		return false
	}
	if patchRequest.SrchTimeEarliest != nil && (roleResponse.SrchTimeEarliest == nil || *patchRequest.SrchTimeEarliest != *roleResponse.SrchTimeEarliest) {
		return false
	}
	if patchRequest.SrchTimeWin != nil && (roleResponse.SrchTimeWin == nil || *patchRequest.SrchTimeWin != *roleResponse.SrchTimeWin) {
		return false
	}
	return true
}
