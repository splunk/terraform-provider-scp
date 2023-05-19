package hec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"net/http"
)

// WaitHecCreate Handles retry logic for POST requests for create lifecycle function
func WaitHecCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createHecRequest v2.CreateHECJSONRequestBody) error {
	waitHecCreateAccepted := wait.GenerateWriteStateChangeConf(HecStatusCreate(ctx, acsClient, stack, createHecRequest))

	rawResp, err := waitHecCreateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for hec (%s) to be created: %s", createHecRequest.Name, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Create response status code for HEC (%s): %d\n", createHecRequest.Name, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for HEC token (%s): %s\n", createHecRequest.Name, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitHecPoll Handles retry logic for polling after POST and DELETE requests for create/delete lifecycle functions
func WaitHecPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string, targetStatus []string, pendingStatus []string) error {
	waitHecState := wait.GenerateReadStateChangeConf(pendingStatus, targetStatus, HecStatusPoll(ctx, acsClient, stack, hecName, targetStatus, pendingStatus))

	_, err := waitHecState.WaitForStateContext(ctx)
	return err
}

// WaitHecRead Handles retry logic for GET requests for the read lifecycle function
func WaitHecRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string) (*v2.HecSpec, error) {
	waitHecRead := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, wait.TargetStatusResourceExists, HecStatusRead(ctx, acsClient, stack, hecName))

	output, err := waitHecRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading hec (%s): %s", hecName, err))
		return nil, err
	}
	hec := output.(*v2.HecSpec)

	return hec, nil
}

// WaitHecUpdate Handles retry logic for PATCH requests for the update lifecycle function
func WaitHecUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchHECJSONRequestBody, hecName string) error {
	waitHecUpdateAccepted := wait.GenerateWriteStateChangeConf(HecStatusUpdate(ctx, acsClient, stack, patchRequest, hecName))

	rawResp, err := waitHecUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for hec (%s) to be updated: %s", hecName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to user that request submitted and update in progress
	tflog.Info(ctx, fmt.Sprintf("Update response status code for hec (%s): %d\n", hecName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for hec (%s): %s\n", hecName, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitVerifyHecUpdate Handles retry logic for GET request for the update lifecycle function to verify that the fields in the
// Hec response match those of the patch request
func WaitVerifyHecUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchHECJSONRequestBody, hecName string) error {
	waitHecUpdateComplete := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, []string{status.UpdatedStatus}, HecStatusVerifyUpdate(ctx, acsClient, stack, patchRequest, hecName))

	_, err := waitHecUpdateComplete.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error confirming hec (%s) has been updated: %s", hecName, err))
		return err
	}

	return nil
}

// WaitHecDelete Handles retry logic for DELETE requests for the delete lifecycle function
func WaitHecDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string) error {
	WaitHecDeleteAccepted := wait.GenerateWriteStateChangeConf(HecStatusDelete(ctx, acsClient, stack, hecName))

	rawResp, err := WaitHecDeleteAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting hec (%s): %s", hecName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to user that request submitted and deletion in progress
	tflog.Info(ctx, fmt.Sprintf("Delete response status code for hec (%s): %d\n", hecName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for hec (%s): %s\n", hecName, resp.Header.Get("X-REQUEST-ID")))
	return nil
}
