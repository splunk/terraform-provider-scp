package indexes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/splunk/terraform-provider-splunkcloud/acs/v2"
	"net/http"
	"time"
)

const (
	CrudDelayTime = 1 * time.Second
	PollDelayTime = 3 * time.Second
	Timeout       = 20 * time.Minute
	PollInterval  = 1 * time.Minute
)

var (
	PendingStatusCRUD          = []string{http.StatusText(429), http.StatusText(403), http.StatusText(500), http.StatusText(503)}
	PendingStatusVerifyCreated = []string{http.StatusText(404), http.StatusText(429), http.StatusText(403), http.StatusText(500), http.StatusText(503)}
	PendingStatusVerifyDeleted = []string{http.StatusText(200), http.StatusText(429), http.StatusText(403), http.StatusText(500), http.StatusText(503)}

	TargetStatusResourceChange  = []string{http.StatusText(202)}
	TargetStatusResourceExists  = []string{http.StatusText(200)}
	TargetStatusResourceDeleted = []string{http.StatusText(404)}
)

func WaitIndexCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createIndexRequest v2.CreateIndexJSONRequestBody) error {
	waitIndexCreateAccepted := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceChange,
		Refresh:      StatusIndexCreate(ctx, acsClient, stack, createIndexRequest),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

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

func WaitIndexPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string, targetStatus []string, pendingStatus []string) error {
	waitIndexCreated := &resource.StateChangeConf{
		Pending:      pendingStatus,
		Target:       targetStatus,
		Refresh:      StatusIndex(ctx, acsClient, stack, indexName, targetStatus, pendingStatus),
		Timeout:      Timeout,
		Delay:        PollDelayTime,
		PollInterval: PollInterval,
	}

	_, err := waitIndexCreated.WaitForStateContext(ctx)
	return err
}

func WaitIndexRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) (*v2.IndexResponse, error) {
	waitIndexRead := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceExists,
		Refresh:      StatusIndexWithIndexResponse(ctx, acsClient, stack, indexName),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

	output, err := waitIndexRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading index (%s): %s", indexName, err))
		return nil, err
	}
	index := output.(*v2.IndexResponse)
	return index, nil
}

func WaitIndexUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchIndexInfoJSONRequestBody, indexName string) error {
	waitIndexUpdateAccepted := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceChange,
		Refresh:      StatusIndexUpdate(ctx, acsClient, stack, patchRequest, indexName),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

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

func WaitIndexConfirmUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, patchRequest v2.PatchIndexInfoJSONRequestBody, indexName string) error {
	waitIndexUpdateAccepted := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       []string{"UPDATED"},
		Refresh:      StatusIndexPollUpdate(ctx, acsClient, stack, patchRequest, indexName),
		Timeout:      Timeout,
		Delay:        PollDelayTime,
		PollInterval: PollInterval,
	}

	_, err := waitIndexUpdateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error confirming index (%s) has been updated: %s", indexName, err))
		return err
	}

	return nil
}

func WaitIndexDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, indexName string) error {
	waitIndexDelete := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceChange,
		Refresh:      StatusIndexDelete(ctx, acsClient, stack, indexName),
		Timeout:      Timeout,
		Delay:        CrudDelayTime, // ToDO check if avoid errors on 1 min delay
		PollInterval: PollInterval,
	}

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
