package indexes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
	"net/http"
)

// WaitIndexCreate Handles retry logic for POST requests for create lifecycle function
func WaitIndexCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createIndexRequest v2.CreateIndexJSONRequestBody) error {
	waitIndexCreateAccepted := wait.GenerateWriteStateChangeConf(IndexStatusCreate(ctx, acsClient, stack, createIndexRequest))

	rawResp, err := waitIndexCreateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for index (%s) to be created: %s", createIndexRequest.Name, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Create response status code for index (%s): %d\n", createIndexRequest.Name, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for index (%s): %s\n", createIndexRequest.Name, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitIndexPoll Handles retry logic for polling after POST and DELETE requests for create/delete lifecycle functions
func WaitIndexPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string, targetStatus []string, pendingStatus []string) error {
	waitIndexCreated := wait.GenerateReadStateChangeConf(pendingStatus, targetStatus, IndexStatusPoll(ctx, acsClient, stack, indexName, targetStatus, pendingStatus))

	_, err := waitIndexCreated.WaitForStateContext(ctx)
	return err
}

// WaitIndexRead Handles retry logic for GET requests for the read lifecycle function
func WaitIndexRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) (*v2.IndexResponse, error) {
	waitIndexRead := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, wait.TargetStatusResourceExists, IndexStatusRead(ctx, acsClient, stack, indexName))

	output, err := waitIndexRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading index (%s): %s", indexName, err))
		return nil, err
	}
	index := output.(*v2.IndexResponse)

	return index, nil
}

// WaitIndexUpdate Handles retry logic for PATCH requests for the update lifecycle function
func WaitIndexUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchIndexInfoJSONRequestBody, indexName string) error {
	waitIndexUpdateAccepted := wait.GenerateWriteStateChangeConf(IndexStatusUpdate(ctx, acsClient, stack, patchRequest, indexName))

	rawResp, err := waitIndexUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for index (%s) to be updated: %s", indexName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to user that request submitted and update in progress
	tflog.Info(ctx, fmt.Sprintf("Update response status code for index (%s): %d\n", indexName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for index (%s): %s\n", indexName, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitVerifyIndexUpdate Handles retry logic for GET request for the update lifecycle function to verify that the fields in the
// index response match those of the patch request
func WaitVerifyIndexUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchIndexInfoJSONRequestBody, indexName string) error {
	waitIndexUpdateAccepted := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, []string{status.UpdatedStatus}, IndexStatusVerifyUpdate(ctx, acsClient, stack, patchRequest, indexName))

	_, err := waitIndexUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error confirming index (%s) has been updated: %s", indexName, err))
		return err
	}

	return nil
}

// WaitIndexDelete Handles retry logic for DELETE requests for the delete lifecycle function
func WaitIndexDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) error {
	waitIndexDelete := wait.GenerateWriteStateChangeConf(IndexStatusDelete(ctx, acsClient, stack, indexName))

	rawResp, err := waitIndexDelete.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error deleting index (%s): %s", indexName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	//Log to user that request submitted and deletion in progress
	tflog.Info(ctx, fmt.Sprintf("Delete response status code for index (%s): %d\n", indexName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for index (%s): %s\n", indexName, resp.Header.Get("X-REQUEST-ID")))
	return nil
}
