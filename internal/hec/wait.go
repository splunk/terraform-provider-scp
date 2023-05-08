package hec

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/splunk/terraform-provider-scp/acs/v2"
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
	PendingStatusCRUD          = []string{http.StatusText(429), http.StatusText(424)}
	PendingStatusVerifyCreated = []string{http.StatusText(404), http.StatusText(429)}
	PendingStatusVerifyDeleted = []string{http.StatusText(200), http.StatusText(429)}

	TargetStatusResourceChange  = []string{http.StatusText(202)}
	TargetStatusResourceExists  = []string{http.StatusText(200)}
	TargetStatusResourceDeleted = []string{http.StatusText(404)}
)

// WaitHecCreate Handles retry logic for POST requests for create lifecycle function
func WaitHecCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createHecRequest v2.CreateHECJSONRequestBody) error {
	waitHecCreateAccepted := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceChange,
		Refresh:      HecStatusCreate(ctx, acsClient, stack, createHecRequest),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

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
	waitHecCreated := &resource.StateChangeConf{
		Pending:      pendingStatus,
		Target:       targetStatus,
		Refresh:      HecStatusPoll(ctx, acsClient, stack, hecName, targetStatus, pendingStatus),
		Timeout:      Timeout,
		Delay:        PollDelayTime,
		PollInterval: PollInterval,
	}

	_, err := waitHecCreated.WaitForStateContext(ctx)
	return err
}

// WaitHecRead Handles retry logic for GET requests for the read lifecycle function
func WaitHecRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, hecName string) (*v2.HecSpec, error) {
	waitHecRead := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceExists,
		Refresh:      HecStatusRead(ctx, acsClient, stack, hecName),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

	output, err := waitHecRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading hec (%s): %s", hecName, err))
		return nil, err
	}
	hec := output.(*v2.HecSpec)

	return hec, nil
}
