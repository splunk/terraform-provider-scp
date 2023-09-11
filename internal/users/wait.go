package users

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"net/http"
)

var (
	TargetStatusResourceExists  = []string{http.StatusText(200)}
	TargetStatusResourceChange  = []string{http.StatusText(200)}
	TargetStatusResourceDeleted = []string{http.StatusText(200)}
)

// WaitUserCreate Handles retry logic for POST requests for create lifecycle function
func WaitUserCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createParams v2.CreateUserParams, createUserRequest v2.CreateUserJSONRequestBody) error {
	waitUserCreateAccepted := wait.GenerateWriteStateChangeConf(UserStatusCreate(ctx, acsClient, stack, createParams, createUserRequest))
	// Override the target status
	waitUserCreateAccepted.Target = TargetStatusResourceExists
	rawResp, err := waitUserCreateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for user (%s) to be created: %s", createUserRequest.Name, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Create response status code for user (%s): %d\n", createUserRequest.Name, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for user (%s): %s\n", createUserRequest.Name, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitUserPoll Handles retry logic for polling after POST and DELETE requests for create/delete lifecycle functions
func WaitUserPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, userName string, targetStatus []string, pendingStatus []string) error {
	waitUserCreated := wait.GenerateReadStateChangeConf(pendingStatus, targetStatus, UserStatusPoll(ctx, acsClient, stack, userName, targetStatus, pendingStatus))

	_, err := waitUserCreated.WaitForStateContext(ctx)
	return err
}

// WaitUserRead Handles retry logic for GET requests for the read lifecycle function
func WaitUserRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, userName string) (*v2.UsersResponse, error) {
	waitUserRead := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, wait.TargetStatusResourceExists, UserStatusRead(ctx, acsClient, stack, userName))

	output, err := waitUserRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading user (%s): %s", userName, err))
		return nil, err
	}
	user := output.(*v2.UsersResponse)

	return user, nil
}

// WaitUserUpdate Handles retry logic for PATCH requests for the update lifecycle function
func WaitUserUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchParams v2.PatchUserParams, patchRequest v2.PatchUserJSONRequestBody, userName string) error {
	waitUserUpdateAccepted := wait.GenerateWriteStateChangeConf(UserStatusUpdate(ctx, acsClient, stack, patchParams, patchRequest, userName))
	waitUserUpdateAccepted.Target = TargetStatusResourceChange
	rawResp, err := waitUserUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for user (%s) to be updated: %s", userName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to user that request submitted and update in progress
	tflog.Info(ctx, fmt.Sprintf("Update response status code for user (%s): %d\n", userName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for user (%s): %s\n", userName, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitVerifyUserUpdate Handles retry logic for GET request for the update lifecycle function to verify that the fields in the
// user response match those of the patch request
func WaitVerifyUserUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchUserJSONRequestBody, userName string) error {
	waitUserUpdateAccepted := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, []string{status.UpdatedStatus}, UserStatusVerifyUpdate(ctx, acsClient, stack, patchRequest, userName))

	_, err := waitUserUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error confirming user (%s) has been updated: %s", userName, err))
		return err
	}

	return nil
}

// WaitUserDelete Handles retry logic for DELETE requests for the delete lifecycle function
func WaitUserDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, userName string) error {
	waitUserDelete := wait.GenerateWriteStateChangeConf(UserStatusDelete(ctx, acsClient, stack, userName))
	waitUserDelete.Target = TargetStatusResourceDeleted
	rawResp, err := waitUserDelete.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting user (%s): %s", userName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to user that request submitted and deletion in progress
	tflog.Info(ctx, fmt.Sprintf("Delete response status code for user (%s): %d\n", userName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for user (%s): %s\n", userName, resp.Header.Get("X-REQUEST-ID")))
	return nil
}
