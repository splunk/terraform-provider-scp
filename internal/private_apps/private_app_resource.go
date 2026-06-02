package privateapps

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/locks"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

const (
	ResourceKey  = "scp_private_app"
	AcsLegalAck  = "acs_legal_ack"
	RetryTimeout = 30 * time.Minute
)

const missingPrivateAppCredentialsMessage = "splunk_username and splunk_password must be configured in the provider to install private apps. Configure provider attributes splunk_username and splunk_password, or set SPLUNK_USERNAME and SPLUNK_PASSWORD, so the provider can create the Splunkbase session and AppInspect login token required by ACS."

func privateAppSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the private app",
			Required:    true,
		},
		"filename": {
			Type:        schema.TypeString,
			Description: "The path to the private app file. The file must be a valid tar.gz archive.",
			Required:    true,
		},
		AcsLegalAck: {
			Type: schema.TypeString,
			Description: "When you install a private app, you must also specify the ACS-Legal-Ack: " +
				"Y parameter to acknowledge your acceptance of any risks involved with the installation of unsupported " +
				"apps on your system, as specified in the Splunk legal disclaimer for app installation, which is " +
				"provided in the ACS OpenAPI 3.0 specification. To review the disclaimer, see Set up the ACS API: " +
				"https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ACSusage#Set_up_the_ACS_API",
			Required: true,
		},
		"pre_vetted": {
			Type:        schema.TypeBool,
			Description: "Whether the private app has been pre-vetted using AppInspect.",
			Required:    true,
		},
		"targets": {
			Type:        schema.TypeSet,
			Description: "List of search head targets to install the application on. Note: TAI functionality must be enabled on the stack and if the app has been previously installed it needs to be imported to Terraform.",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func ResourcePrivateApp() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Private App. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageApps for more latest, detailed information on attribute requirements and the ACS API.",
		CreateContext: resourcePrivateAppCreate,
		UpdateContext: resourcePrivateAppUpdate,
		ReadContext:   resourcePrivateAppRead,
		DeleteContext: resourcePrivateAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: privateAppSchema(),
	}
}

func resourcePrivateAppCreate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	tflog.Info(ctx, "[BETA] Private Apps: This feature is in beta release.")
	acsProvider := m.(*client.ACSProvider)
	if !resourceData.Get("pre_vetted").(bool) {
		return diag.Errorf("App must be pre-vetted before it can be installed. Set 'pre_vetted' to true manually after vetting the app or use scp_app_validation data source to automate the process.")
	}
	if acsProvider.AppInspectClient == nil {
		return diag.Errorf(missingPrivateAppCredentialsMessage)
	}
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack
	splunkbase := false

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "create")
	defer unlock()

	ACSLegalAck := resourceData.Get("acs_legal_ack").(string)
	installParams := v2.InstallAppVictoriaParams{
		Splunkbase:  &splunkbase,
		ACSLegalAck: &ACSLegalAck,
	}

	targetsRaw, targetsOk := resourceData.GetOk("targets")
	targetInstalls := []client.TargetFields{}
	if targetsOk && len(targetsRaw.(*schema.Set).List()) > 0 {
		targets := targetsRaw.(*schema.Set).List()

		for _, targetRaw := range targets {
			targetInstall, err := client.GetTargetInstall(ctx, targetRaw.(string), stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			targetInstalls = append(targetInstalls, targetInstall)
		}
		localScope := "local"
		installParams.Scope = &localScope
	} else {
		targetInstalls = append(targetInstalls, client.TargetFields{
			Client: acsClient,
			Stack:  stack,
		})
	}

	appID := ""
	for _, targetInstall := range targetInstalls {
		retriesOfBadPackage := 0

		err := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
			fileData, errFileRead := os.ReadFile(resourceData.Get("filename").(string))
			if errFileRead != nil {
				return resource.NonRetryableError(errFileRead)
			}
			body := bytes.NewReader(fileData)
			output, err := WaitAppCreate(ctx, targetInstall.Client, targetInstall.Stack, installParams, body)
			if err != nil {
				if errors.IsConflictError(err.Err) {
					return resource.NonRetryableError(err.Err)
				}
				if strings.Contains(err.Err.Error(), "503") {
					return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %v", err.Err))
				}
				if strings.Contains(err.Err.Error(), "Extract app information from the package failed") || strings.Contains(err.Err.Error(), "app package not found in the request") && retriesOfBadPackage < 3 {
					retriesOfBadPackage++
					return resource.RetryableError(fmt.Errorf("received 'Extract app information from the package failed' error, retrying: %v", err.Err))
				}
				if err.Retryable {
					return resource.RetryableError(fmt.Errorf("retryable error occurred: %v", err.Err))
				}
				return resource.NonRetryableError(err.Err)
			}
			appID = *output.AppID
			return nil
		})

		if err != nil {
			if errors.IsConflictError(err) {
				tflog.Info(ctx, "App (%s) already exists, if you want to update it, change app's version")
				resourceData.SetId(appID)
			} else {
				return diag.Errorf("Error submitting request for app to be created. %v", err)
			}
		}

		err = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
			err := WaitAppPoll(ctx, targetInstall.Client, targetInstall.Stack, appID, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
			if err != nil {
				if strings.Contains(err.Error(), "503") {
					return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
				}
				if strings.Contains(err.Error(), "404") {
					return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
				}
				return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be created: %s", resourceData.Get("name").(string), err))
			}
			return nil
		})
		if err != nil {
			return diag.FromErr(err)
		}
	}

	if appID == "" {
		appID = resourceData.Get("name").(string)
	}

	resourceData.SetId(appID)
	return nil
}

func resourcePrivateAppRead(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	acsProvider := m.(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	targetReads := []client.TargetFields{}
	targetsRaw, targetsOk := resourceData.GetOk("targets")
	if targetsOk && len(targetsRaw.(*schema.Set).List()) > 0 {
		for _, targetRaw := range targetsRaw.(*schema.Set).List() {
			targetRead, err := client.GetTargetInstall(ctx, targetRaw.(string), stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			targetReads = append(targetReads, targetRead)
		}
	} else {
		targetReads = append(targetReads, client.TargetFields{
			Client: acsClient,
			Stack:  stack,
		})
	}

	appName := resourceData.Id()
	for _, targetRead := range targetReads {
		_, err := WaitAppRead(ctx, targetRead.Client, targetRead.Stack, appName)
		if err != nil {
			if stateErr, ok := err.(*resource.UnexpectedStateError); ok && strings.Contains(stateErr.LastError.Error(), "404-app-not-found") {
				return nil
			}
			return diag.Errorf("Error reading app (%s): %s", appName, err)
		}
	}

	return nil
}

func resourcePrivateAppDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "[BETA] Private Apps: This feature is in beta release.")

	acsProvider := m.(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "delete")
	defer unlock()

	uninstallParams := v2.UninstallAppVictoriaParams{}
	targetsRaw, targetsOk := resourceData.GetOk("targets")
	targetDeletes := []client.TargetFields{}
	if targetsOk && len(targetsRaw.(*schema.Set).List()) > 0 {
		for _, targetRaw := range targetsRaw.(*schema.Set).List() {
			targetDelete, err := client.GetTargetInstall(ctx, targetRaw.(string), stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			targetDeletes = append(targetDeletes, targetDelete)
		}
		localScope := "local"
		uninstallParams.Scope = &localScope
	} else {
		targetDeletes = append(targetDeletes, client.TargetFields{
			Client: acsClient,
			Stack:  stack,
		})
	}

	appName := resourceData.Id()
	for _, targetDelete := range targetDeletes {
		retryErr := resource.RetryContext(ctx, 2*RetryTimeout, func() *resource.RetryError {
			err := WaitAppDelete(ctx, targetDelete.Client, targetDelete.Stack, appName, uninstallParams)
			if err != nil {
				if strings.Contains(err.Error(), "503") {
					return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})
		if retryErr != nil {
			return diag.Errorf("Error deleting app (%s): %s", appName, retryErr)
		}

		retryErr = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
			err := WaitAppPoll(ctx, targetDelete.Client, targetDelete.Stack, appName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
			if err != nil {
				if strings.Contains(err.Error(), "503") {
					return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
				}
				return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
			}
			return nil
		})
		if retryErr != nil {
			return diag.Errorf("Error waiting for app (%s) to be deleted: %s", appName, retryErr)
		}
	}

	return nil
}

func resourcePrivateAppUpdate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "[BETA] Private Apps: This feature is in beta release.")
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(*client.ACSProvider)
	if !resourceData.Get("pre_vetted").(bool) {
		return diag.Errorf("App must be pre-vetted before it can be installed. Set 'pre_vetted' to true manually after vetting the app or use scp_app_validation data source to automate the process.")
	}
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "update")
	defer unlock()

	fileData, err := os.ReadFile(resourceData.Get("filename").(string))
	if err != nil {
		return diag.FromErr(err)
	}

	acsLegalAck := resourceData.Get("acs_legal_ack").(string)
	isSplunkbase := false
	appName := resourceData.Get("name").(string)

	oldTargetsRaw, newTargetsRaw := resourceData.GetChange("targets")
	oldTargets := utils.SetToStrings(oldTargetsRaw)
	newTargets := utils.SetToStrings(newTargetsRaw)
	oldHasTargets := len(oldTargets) > 0
	newHasTargets := len(newTargets) > 0
	targetsChanged := resourceData.HasChange("targets")

	switch {
	case targetsChanged && !oldHasTargets && newHasTargets:
		// No targets -> targets: install on selected targets
		if acsProvider.AppInspectClient == nil {
			return diag.Errorf(missingPrivateAppCredentialsMessage)
		}
		tflog.Warn(ctx, "Targets were not previously tracked. Only installs will be performed. To remove from specific targets, migrate the resource with targets.")
		for target := range newTargets {
			targetInstall, targetErr := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if targetErr != nil {
				return diag.FromErr(targetErr)
			}
			localScope := "local"
			_, targetErr = createPrivateApp(ctx, targetInstall, v2.InstallAppVictoriaParams{
				Splunkbase:  &isSplunkbase,
				ACSLegalAck: &acsLegalAck,
				Scope:       &localScope,
			}, fileData, appName)
			if targetErr != nil {
				return diag.Errorf("Error submitting request for app to be created. %v", targetErr)
			}
		}
	case targetsChanged && oldHasTargets && !newHasTargets:
		// Targets -> no targets: install the app everywhere
		if acsProvider.AppInspectClient == nil {
			return diag.Errorf(missingPrivateAppCredentialsMessage)
		}
		_, createErr := createPrivateApp(ctx, client.TargetFields{
			Client: acsClient,
			Stack:  stack,
		}, v2.InstallAppVictoriaParams{
			Splunkbase:  &isSplunkbase,
			ACSLegalAck: &acsLegalAck,
		}, fileData, appName)
		if createErr != nil {
			return diag.Errorf("Error submitting request for app to be created. %v", createErr)
		}
	case targetsChanged && oldHasTargets && newHasTargets:
		// Targets -> targets: remove deleted, add new
		for target := range newTargets {
			if _, ok := oldTargets[target]; ok {
				continue
			}
			if acsProvider.AppInspectClient == nil {
				return diag.Errorf(missingPrivateAppCredentialsMessage)
			}
			targetInstall, targetErr := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if targetErr != nil {
				return diag.FromErr(targetErr)
			}
			localScope := "local"
			_, targetErr = createPrivateApp(ctx, targetInstall, v2.InstallAppVictoriaParams{
				Splunkbase:  &isSplunkbase,
				ACSLegalAck: &acsLegalAck,
				Scope:       &localScope,
			}, fileData, appName)
			if targetErr != nil {
				return diag.Errorf("Error submitting request for app to be created. %v", targetErr)
			}
		}
		for target := range oldTargets {
			if _, ok := newTargets[target]; ok {
				continue
			}
			targetDelete, targetErr := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if targetErr != nil {
				return diag.FromErr(targetErr)
			}
			if targetErr = deleteOnTarget(ctx, targetDelete, resourceData.Id()); targetErr != nil {
				return diag.Errorf("Error deleting app (%s): %s", resourceData.Id(), targetErr)
			}
		}
	}

	if updateVersion := resourceData.HasChange("filename"); !updateVersion {
		return nil
	}
	if acsProvider.AppInspectClient == nil {
		return diag.Errorf(missingPrivateAppCredentialsMessage)
	}

	updateTargetsRaw, updateTargetsOk := resourceData.GetOk("targets")
	updateTargets := []client.TargetFields{}
	if updateTargetsOk && len(updateTargetsRaw.(*schema.Set).List()) > 0 {
		for _, targetRaw := range updateTargetsRaw.(*schema.Set).List() {
			target := strings.TrimSpace(targetRaw.(string))
			if target == "" {
				return diag.Errorf("targets must not contain empty entries")
			}
			targetInstall, err := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			updateTargets = append(updateTargets, targetInstall)
		}
	} else {
		updateTargets = append(updateTargets, client.TargetFields{
			Client: acsClient,
			Stack:  stack,
		})
	}

	isTargetScoped := updateTargetsOk && len(updateTargetsRaw.(*schema.Set).List()) > 0
	for _, targetUpdate := range updateTargets {
		installParams := v2.InstallAppVictoriaParams{
			Splunkbase:  &isSplunkbase,
			ACSLegalAck: &acsLegalAck,
		}
		if isTargetScoped {
			localScope := "local"
			installParams.Scope = &localScope
		}

		if _, err := createPrivateApp(ctx, targetUpdate, installParams, fileData, appName); err != nil {
			return diag.Errorf("Error updating app (%s): %s", resourceData.Id(), err)
		}
	}

	return nil
}

func createPrivateApp(ctx context.Context, targetInstall client.TargetFields, installParams v2.InstallAppVictoriaParams, fileData []byte, fallbackID string) (string, error) {
	appID := ""
	retriesOfBadPackage := 0

	err := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		body := bytes.NewReader(fileData)
		output, err := WaitAppCreate(ctx, targetInstall.Client, targetInstall.Stack, installParams, body)
		if err != nil {
			if errors.IsConflictError(err.Err) {
				existingApp, readErr := WaitAppRead(ctx, targetInstall.Client, targetInstall.Stack, fallbackID)
				if readErr == nil {
					appID = appIdentifier(existingApp, fallbackID)
				} else if appID == "" {
					appID = fallbackID
				}
				tflog.Info(ctx, "App already installed on target; skipping install and keeping Terraform state in sync",
					map[string]any{"stack": string(targetInstall.Stack), "app": fallbackID})
				return nil
			}
			if strings.Contains(err.Err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %v", err.Err))
			}
			if (strings.Contains(err.Err.Error(), "Extract app information from the package failed") || strings.Contains(err.Err.Error(), "app package not found in the request")) && retriesOfBadPackage < 3 {
				retriesOfBadPackage++
				return resource.RetryableError(fmt.Errorf("received 'Extract app information from the package failed' error, retrying: %v", err.Err))
			}
			if err.Retryable {
				return resource.RetryableError(fmt.Errorf("retryable error occurred: %v", err.Err))
			}
			return resource.NonRetryableError(err.Err)
		}
		appID = appIdentifier(output, fallbackID)
		return nil
	})

	if appID == "" {
		appID = fallbackID
	}

	if err != nil {
		return appID, err
	}

	err = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		err := WaitAppPoll(ctx, targetInstall.Client, targetInstall.Stack, appID, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "404") {
				return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be created: %s", fallbackID, err))
		}
		return nil
	})

	return appID, err
}

func deleteOnTarget(ctx context.Context, targetDelete client.TargetFields, appName string) error {
	localScope := "local"
	deleteParams := v2.UninstallAppVictoriaParams{
		Scope: &localScope,
	}

	retryErr := resource.RetryContext(ctx, 2*RetryTimeout, func() *resource.RetryError {
		err := WaitAppDelete(ctx, targetDelete.Client, targetDelete.Stack, appName, deleteParams)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	return resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		err := WaitAppPoll(ctx, targetDelete.Client, targetDelete.Stack, appName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
		}
		return nil
	})
}

func appIdentifier(app *v2.App, fallback string) string {
	if app != nil {
		if app.AppID != nil && *app.AppID != "" {
			return *app.AppID
		}
		if strings.TrimSpace(app.Name) != "" {
			return app.Name
		}
	}
	return fallback
}
