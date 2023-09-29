package roles

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"net/http"
)

// For all CRUD operations on synchronous resources we expect a target status of 200.
var (
	TargetStatusResourceExists  = []string{http.StatusText(200)}
	TargetStatusResourceChange  = []string{http.StatusText(200)}
	TargetStatusResourceDeleted = []string{http.StatusText(200)}
)

// WaitRoleCreate Handles retry logic for POST requests for create lifecycle function
func WaitRoleCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createParams v2.CreateRoleParams, createRoleRequest v2.CreateRoleJSONRequestBody) error {
	waitRoleCreateAccepted := wait.GenerateWriteStateChangeConf(RoleStatusCreate(ctx, acsClient, stack, createParams, createRoleRequest))
	// Override the target status
	waitRoleCreateAccepted.Target = TargetStatusResourceExists
	rawResp, err := waitRoleCreateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for role (%s) to be created: %s", createRoleRequest.Name, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to role that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Create response status code for role (%s): %d\n", createRoleRequest.Name, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for role (%s): %s\n", createRoleRequest.Name, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitRoleRead Handles retry logic for GET requests for the read lifecycle function
func WaitRoleRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, roleName string) (*v2.RolesResponse, error) {
	waitRoleRead := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, wait.TargetStatusResourceExists, RoleStatusRead(ctx, acsClient, stack, roleName))

	output, err := waitRoleRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading role (%s): %s", roleName, err))
		return nil, err
	}
	role := output.(*v2.RolesResponse)

	return role, nil
}

// WaitRoleUpdate Handles retry logic for PATCH requests for the update lifecycle function
func WaitRoleUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchParams v2.PatchRoleInfoParams, patchRequest v2.PatchRoleInfoJSONRequestBody, roleName string) error {
	waitRoleUpdateAccepted := wait.GenerateWriteStateChangeConf(RoleStatusUpdate(ctx, acsClient, stack, patchParams, patchRequest, roleName))
	waitRoleUpdateAccepted.Target = TargetStatusResourceChange
	rawResp, err := waitRoleUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for role (%s) to be updated: %s", roleName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to role that request submitted and update in progress
	tflog.Info(ctx, fmt.Sprintf("Update response status code for role (%s): %d\n", roleName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for role (%s): %s\n", roleName, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitVerifyRoleUpdate Handles retry logic for GET request for the update lifecycle function to verify that the fields in the
// role response match those of the patch request
func WaitVerifyRoleUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchRoleInfoJSONRequestBody, roleName string) error {
	waitRoleUpdateAccepted := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, []string{status.UpdatedStatus}, RoleStatusVerifyUpdate(ctx, acsClient, stack, patchRequest, roleName))

	_, err := waitRoleUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error confirming role (%s) has been updated: %s", roleName, err))
		return err
	}

	return nil
}

// WaitRoleDelete Handles retry logic for DELETE requests for the delete lifecycle function
func WaitRoleDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, roleName string) error {
	waitRoleDelete := wait.GenerateWriteStateChangeConf(RoleStatusDelete(ctx, acsClient, stack, roleName))
	waitRoleDelete.Target = TargetStatusResourceDeleted
	rawResp, err := waitRoleDelete.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting role (%s): %s", roleName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to role that request submitted and deletion in progress
	tflog.Info(ctx, fmt.Sprintf("Delete response status code for role (%s): %d\n", roleName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for role (%s): %s\n", roleName, resp.Header.Get("X-REQUEST-ID")))
	return nil
}
