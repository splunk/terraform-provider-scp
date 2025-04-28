package ipallowlists

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
)

const (
	CrudDelayTime = 1 * time.Second
	PollDelayTime = 3 * time.Second
	Timeout       = 20 * time.Minute
	PollInterval  = 1 * time.Minute
)

var (
	PendingStatusCRUD          = []string{http.StatusText(http.StatusTooManyRequests)}
	PendingStatusVerifyCreated = []string{http.StatusText(http.StatusTooManyRequests)}
	PendingStatusVerifyDeleted = []string{http.StatusText(200), http.StatusText(http.StatusTooManyRequests)}

	TargetStatusResourceChange  = []string{http.StatusText(200)}
	TargetStatusResourceExists  = []string{http.StatusText(200)}
	TargetStatusResourceDeleted = []string{http.StatusText(404)}
)

// WaitIPAllowlistCreate Handles retry logic for POST requests for create lifecycle function
func WaitIPAllowlistCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, feature v2.Feature, newSubnets []string) error {
	waitIPAllowlistCreateAccepted := &resource.StateChangeConf{
		Target:       TargetStatusResourceChange,
		Refresh:      IPAllowlistStatusCreate(ctx, acsClient, stack, feature, newSubnets),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

	rawResp, err := waitIPAllowlistCreateAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for ip allowlist (%s) to be created: %s", feature, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Create response status code for ip allowlist (%s): %d\n", feature, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for IP allowlist (%s): %s\n", feature, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

// WaitIPAllowlistRead Handles retry logic for GET requests for the read lifecycle function
func WaitIPAllowlistRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, feature string) ([]string, error) {
	waitIPAllowlistRead := &resource.StateChangeConf{
		Pending:      PendingStatusCRUD,
		Target:       TargetStatusResourceExists,
		Refresh:      IPAllowlistStatusRead(ctx, acsClient, stack, feature),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

	output, err := waitIPAllowlistRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading ip allowlist (%s): %s", feature, err))
		return nil, err
	}
	subnets := output.([]string)

	return subnets, nil
}

// WaitIPAllowlistDelete Handles retry logic for POST requests for delete lifecycle function
func WaitIPAllowlistDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, feature v2.Feature, oldSubnets []string) error {
	waitIPAllowlistDeleteAccepted := &resource.StateChangeConf{
		Target:       TargetStatusResourceChange,
		Refresh:      IPAllowlistStatusDelete(ctx, acsClient, stack, feature, oldSubnets),
		Timeout:      Timeout,
		Delay:        CrudDelayTime,
		PollInterval: PollInterval,
	}

	rawResp, err := waitIPAllowlistDeleteAccepted.WaitForStateContext(ctx)
	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for ip allowlist (%s) to be deleted: %s", feature, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Delete response status code for ip allowlist (%s): %d\n", feature, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for IP allowlist (%s): %s\n", feature, resp.Header.Get("X-REQUEST-ID")))

	return nil
}
