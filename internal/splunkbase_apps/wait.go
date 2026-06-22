package splunkbaseapps

import (
	"context"
	"fmt"
	"io"
	"net/http"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

// WaitAppCreate Handles retry logic for POST requests for create lifecycle function
func WaitAppCreate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, createAppRequest v2.InstallAppVictoriaParams, body io.Reader) *resource.RetryError {
	waitAppCreateAccepted := wait.GenerateWriteStateChangeConf(AppStatusCreate(ctx, acsClient, stack, createAppRequest, body))
	rawResp, err := waitAppCreateAccepted.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for App (%s) to be created", err))
		return resource.NonRetryableError(err)
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Create response status code for App: %d\n", resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID (%s):", resp.Header.Get("X-REQUEST-ID")))

	return nil
}

func WaitAppRead(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName string) (*v2.App, error) {
	waitAppRead := wait.GenerateReadStateChangeConf(wait.PendingStatusCRUD, wait.TargetStatusResourceExists, AppStatusRead(ctx, acsClient, stack, v2.AppName(appName)))
	output, err := waitAppRead.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error reading App %s: %s", appName, err))
		return nil, err
	}

	app := output.(*v2.App)
	return app, nil
}

func WaitAppDelete(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName string, params v2.UninstallAppVictoriaParams) error {
	waitAppDelete := wait.GenerateDeleteStateChangeConf(AppStatusDelete(ctx, acsClient, stack, v2.AppName(appName), params))
	rawResp, err := waitAppDelete.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for App to be deleted (%s)", err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and deletion in progress
	tflog.Info(ctx, fmt.Sprintf("Delete response status code for App: %d\n", resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID (%s):", resp.Header.Get("X-REQUEST-ID")))

	return nil
}

func WaitAppUpdate(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName string, createAppRequest v2.PatchAppVictoriaParams, body io.Reader) error {
	waitAppUpdateAccepted := wait.GenerateWriteStateChangeConf(AppStatusUpdate(ctx, acsClient, stack, v2.AppName(appName), createAppRequest, body))
	rawResp, err := waitAppUpdateAccepted.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error submitting request for App (%s) to be updated: %s", appName, err))
		return err
	}

	resp := rawResp.(*http.Response)

	// Log to user that request submitted and creation in progress
	tflog.Info(ctx, fmt.Sprintf("Update response status code for App (%s): %d\n", appName, resp.StatusCode))
	tflog.Info(ctx, fmt.Sprintf("ACS Request ID for App (%s): %s\n", appName, resp.Header.Get("X-REQUEST-ID")))

	return nil
}

func WaitAppPoll(ctx context.Context, acsClient v2.ClientInterface, stack v2.Stack, appName string, targetStatus []string, pendingStatus []string) error {
	waitAppPoll := wait.GenerateReadStateChangeConf(pendingStatus, targetStatus, AppStatusPoll(ctx, acsClient, stack, v2.AppName(appName), targetStatus, pendingStatus))
	_, err := waitAppPoll.WaitForStateContext(ctx)

	if err != nil {
		tflog.Error(ctx, fmt.Sprintf("Error polling App (%s): %s", appName, err))
		return err
	}

	return nil
}
