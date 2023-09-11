package users

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"io"
	"net/http"
)

var GeneralRetryableStatusCodes = map[int]string{
	http.StatusTooManyRequests: http.StatusText(429),
}

// UserStatusCreate returns StateRefreshFunc that makes POST request and checks if response is accepted
func UserStatusCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createParams v2.CreateUserParams, createUserRequest v2.CreateUserJSONRequestBody) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {
		resp, err := acsClient.CreateUser(ctx, stack, &createParams, createUserRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		return status.ProcessResponse(resp, wait.TargetStatusResourceExists, wait.PendingStatusCRUD)
	}
}

// UserStatusPoll returns StateRefreshFunc that makes GET request and checks if response is desired target (200 for create and 404 for delete)
func UserStatusPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, userName string, targetStatus []string, pendingStatus []string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeUser(ctx, stack, v2.UserName(userName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, targetStatus, pendingStatus)
	}
}

// UserStatusRead returns StateRefreshFunc that makes GET request, checks if request was successful, and returns user response
func UserStatusRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, userName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeUser(ctx, stack, v2.UserName(userName))
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

		var user v2.UsersResponse
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &user); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
		}
		status := http.StatusText(resp.StatusCode)
		return &user, status, nil
	}
}

// UserStatusDelete returns StateRefreshFunc that makes DELETE request and checks if request was accepted
func UserStatusDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, userName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DeleteUser(ctx, stack, v2.UserName(userName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, wait.TargetStatusResourceExists, wait.PendingStatusCRUD)
	}
}

// UserStatusUpdate returns StateRefreshFunc that makes PATCH request and checks if request was successful
func UserStatusUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchParams v2.PatchUserParams, patchUserRequest v2.PatchUserJSONRequestBody, userName string) resource.StateRefreshFunc {
	return func() (interface{}, string, error) {

		resp, err := acsClient.PatchUser(ctx, stack, v2.UserName(userName), &patchParams, patchUserRequest)
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()

		return status.ProcessResponse(resp, TargetStatusResourceChange, wait.PendingStatusCRUD)
	}
}

// UserStatusVerifyUpdate returns a StateRefreshFunc that makes a GET request and checks to see if the user fields matches those in patch request
func UserStatusVerifyUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchUserJSONRequestBody, userName string) resource.StateRefreshFunc {
	return func() (any, string, error) {
		resp, err := acsClient.DescribeUser(ctx, stack, v2.UserName(userName))
		if err != nil {
			return nil, "", &resource.UnexpectedStateError{LastError: err}
		}
		defer resp.Body.Close()
		bodyBytes, _ := io.ReadAll(resp.Body)

		if _, ok := GeneralRetryableStatusCodes[resp.StatusCode]; !ok && resp.StatusCode != 200 {
			return nil, http.StatusText(resp.StatusCode), &resource.UnexpectedStateError{LastError: errors.New(string(bodyBytes))}
		}

		var user UserInfo
		updateComplete := false
		if resp.StatusCode == 200 {
			if err = json.Unmarshal(bodyBytes, &user); err != nil {
				return nil, "", &resource.UnexpectedStateError{LastError: err}
			}
			updateComplete = VerifyUserUpdate(patchRequest, user)
		}

		var statusText string
		if updateComplete {
			statusText = status.UpdatedStatus
			return &user, statusText, nil
		} else {
			statusText = http.StatusText(resp.StatusCode)
			return nil, statusText, nil
		}
	}
}

// VerifyUserUpdate is a helper to verify that the fields in patch request match fields in the user response
func VerifyUserUpdate(patchRequest v2.PatchUserJSONRequestBody, user UserInfo) bool {
	if patchRequest.Roles != nil && !utils.IsSpliceEqual(patchRequest.Roles, user.Roles) {
		return false
	}
	if patchRequest.FullName != nil && *patchRequest.FullName != *user.FullName {
		return false
	}
	if patchRequest.DefaultApp != nil && *patchRequest.DefaultApp != *user.DefaultApp {
		return false
	}
	if patchRequest.Email != nil && *patchRequest.Email != *user.Email {
		return false
	}
	return true
}
