package splunkbaseapps

import (
	"context"
	"fmt"
	"net/url"
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
	ResourceKey     = "scp_splunkbase_app"
	splunkbaseID    = "splunkbase_id"
	AcsLicensingAck = "acs_licensing_ack"
	RetryTimeout    = 30 * time.Minute
)

const missingSplunkbaseAppCredentialsMessage = "splunk_username and splunk_password must be configured in the provider to install Splunkbase apps. Configure provider attributes splunk_username and splunk_password, or set SPLUNK_USERNAME and SPLUNK_PASSWORD, so the provider can create the Splunkbase session required by ACS."

func splunkbaseAppSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Description: "The name of the Splunkbase app",
			Required:    true,
		},
		splunkbaseID: {
			Type:        schema.TypeString,
			Description: "The ID of the Splunkbase app",
			Required:    true,
		},
		"version": {
			Type:        schema.TypeString,
			Description: "The version of the Splunkbase app",
			Required:    true,
		},
		AcsLicensingAck: {
			Type:        schema.TypeString,
			Description: "The app's third-party license URL. The license URL is available under 'Licensing' on the Splunkbase download page for the app.",
			Required:    true,
		},
		"targets": {
			Type:        schema.TypeSet,
			Description: "List of roles to install the application on. Note: TAI functionality must be enabled on the stack and if the app has been previously installed it needs to be imported to the terraform.",
			Optional:    true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
		},
	}
}

func ResourceSplunkbaseApp() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description:   "Splunkbase App. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/9.3.2408/Config/ManageSplunkbaseApps for more latest, detailed information on attribute requirements and the ACS API.",
		CreateContext: resourceSplunkbaseAppCreate,
		UpdateContext: resourceSplunkbaseAppUpdate,
		ReadContext:   resourceSplunkbaseAppRead,
		DeleteContext: resourceSplunkbaseAppDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},
		Schema: splunkbaseAppSchema(),
	}
}

func resourceSplunkbaseAppCreate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	tflog.Info(ctx, "[BETA] Splunkbase Apps: This feature is in beta release.")
	acsProvider := m.(*client.ACSProvider)
	if acsProvider.AppInspectClient == nil {
		return diag.Errorf(missingSplunkbaseAppCredentialsMessage)
	}
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack
	splunkbase := true

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "create")
	defer unlock()

	ACSLicensingAck := resourceData.Get("acs_licensing_ack").(string)
	installParams := v2.InstallAppVictoriaParams{
		Splunkbase:      &splunkbase,
		ACSLicensingAck: &ACSLicensingAck,
	}

	version, ok := resourceData.GetOk("version")
	if !ok || version.(string) == "" {
		return diag.Errorf("version must be provided")
	}
	name, ok := resourceData.GetOk("name")
	if !ok || name.(string) == "" {
		return diag.Errorf("name must be provided")
	}

	data := url.Values{}
	data.Set("name", name.(string))
	data.Set("version", version.(string))
	splunkbaseIDParam, ok := resourceData.GetOk("splunkbase_id")
	if !ok || splunkbaseIDParam.(string) == "" {
		return diag.Errorf("splunkbase_id must be provided")
	}

	data.Set("splunkbaseID", splunkbaseIDParam.(string))
	encoded := data.Encode()

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

	for _, targetInstall := range targetInstalls {
		err := installApp(ctx, targetInstall, resourceData.Get("name").(string), ACSLicensingAck, encoded, installParams.Scope)
		if err != nil {
			if errors.IsConflictError(err) {
				tflog.Info(ctx, "App (%s) already exists, if you want to update it, change app's version")
				resourceData.SetId(resourceData.Get("name").(string))
			} else {
				return diag.Errorf("Error submitting request for app to be created. %v", err)
			}
		}
	}
	resourceData.SetId(resourceData.Get("name").(string))
	return nil
}

func resourceSplunkbaseAppRead(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	acsProvider := m.(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	targetReads := []client.TargetFields{}
	targetsRaw, targetsOk := resourceData.GetOk("targets")
	if targetsOk && len(targetsRaw.(*schema.Set).List()) > 0 {
		targets := targetsRaw.(*schema.Set).List()
		for _, targetRaw := range targets {
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
		err := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
			_, err := WaitAppRead(ctx, targetRead.Client, targetRead.Stack, appName)
			if err != nil {
				if stateErr, ok := err.(*resource.UnexpectedStateError); ok && strings.Contains(stateErr.LastError.Error(), "404-app-not-found") {
					return resource.NonRetryableError(err)
				}
				if strings.Contains(err.Error(), "503") {
					return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
				}
				if strings.Contains(err.Error(), "400") {
					return resource.RetryableError(fmt.Errorf("received 400 error, retrying: %w", err))
				}
				if strings.Contains(err.Error(), "404") {
					return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
				}
				return resource.NonRetryableError(err)
			}
			return nil
		})

		if err != nil {
			if stateErr, ok := err.(*resource.UnexpectedStateError); ok && strings.Contains(stateErr.LastError.Error(), "404-app-not-found") {
				return nil
			}
			return diag.Errorf("Error reading app (%s): %s", appName, err)
		}
	}

	return nil
}

func resourceSplunkbaseAppDelete(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "[BETA] Splunkbase Apps: This feature is in beta release.")

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
		targets := targetsRaw.(*schema.Set).List()
		for _, targetRaw := range targets {
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

func resourceSplunkbaseAppUpdate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "[BETA] Splunkbase Apps: This feature is in beta release.")
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "update")
	defer unlock()

	appName := resourceData.Id()

	ACSLicensingAck := resourceData.Get("acs_licensing_ack").(string)

	installParams := v2.PatchAppVictoriaParams{
		ACSLicensingAck: ACSLicensingAck,
	}

	data := url.Values{}
	data.Set("name", resourceData.Get("name").(string))
	data.Set("version", resourceData.Get("version").(string))
	data.Set("splunkbaseID", resourceData.Get("splunkbase_id").(string))

	encoded := data.Encode()

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
			return diag.Errorf(missingSplunkbaseAppCredentialsMessage)
		}
		tflog.Warn(ctx, "Targets were not previously tracked. Only installs will be performed. To remove from specific targets, migrate the resource with targets.")
		for target := range newTargets {
			targetInstall, err := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := installOnTarget(ctx, targetInstall, appName, ACSLicensingAck, encoded); err != nil {
				return diag.Errorf("Error submitting request for app to be created. %v", err)
			}
		}
	case targetsChanged && oldHasTargets && !newHasTargets:
		// Targets -> no targets: install the app everywhere
		if acsProvider.AppInspectClient == nil {
			return diag.Errorf(missingSplunkbaseAppCredentialsMessage)
		}
		if err := installApp(ctx, client.TargetFields{
			Client: acsClient,
			Stack:  stack,
		}, appName, ACSLicensingAck, encoded, nil); err != nil {
			return diag.Errorf("Error submitting request for app to be created. %v", err)
		}
	case targetsChanged && oldHasTargets && newHasTargets:
		// Targets -> targets: remove deleted, add new
		for target := range newTargets {
			if _, ok := oldTargets[target]; ok {
				continue
			}
			if acsProvider.AppInspectClient == nil {
				return diag.Errorf(missingSplunkbaseAppCredentialsMessage)
			}
			targetInstall, err := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := installOnTarget(ctx, targetInstall, appName, ACSLicensingAck, encoded); err != nil {
				return diag.Errorf("Error submitting request for app to be created. %v", err)
			}
		}
		for target := range oldTargets {
			if _, ok := newTargets[target]; ok {
				continue
			}
			targetInstall, err := client.GetTargetInstall(ctx, target, stack, acsProvider)
			if err != nil {
				return diag.FromErr(err)
			}
			if err := deleteOnTarget(ctx, targetInstall, appName); err != nil {
				return diag.Errorf("Error deleting app (%s): %s", appName, err)
			}
		}
	}

	versionChanged := resourceData.HasChange("version")
	licensingChanged := resourceData.HasChange(AcsLicensingAck)
	if !versionChanged && !licensingChanged {
		return nil
	}
	if acsProvider.AppInspectClient == nil {
		return diag.Errorf(missingSplunkbaseAppCredentialsMessage)
	}

	updateClient := acsClient
	updateStack := stack
	updateTargetsRaw, updateTargetsOk := resourceData.GetOk("targets")
	if updateTargetsOk && len(updateTargetsRaw.(*schema.Set).List()) > 0 {
		firstTarget := strings.TrimSpace(updateTargetsRaw.(*schema.Set).List()[0].(string))
		if firstTarget == "" {
			return diag.Errorf("targets must not contain empty entries")
		}
		targetInstall, err := client.GetTargetInstall(ctx, firstTarget, stack, acsProvider)
		if err != nil {
			return diag.FromErr(err)
		}
		updateClient = targetInstall.Client
		updateStack = targetInstall.Stack
	}

	if err := updateApp(ctx, updateClient, updateStack, appName, installParams, encoded, resourceData.Get("version").(string), versionChanged); err != nil {
		return diag.Errorf("Error updating app (%s): %s", appName, err)
	}

	return nil
}

func installApp(ctx context.Context, targetInstall client.TargetFields, appName string, ACSLicensingAck string, encoded string, scope *string) error {
	isSplunkbase := true
	createParams := v2.InstallAppVictoriaParams{
		Splunkbase:      &isSplunkbase,
		ACSLicensingAck: &ACSLicensingAck,
		Scope:           scope,
	}

	retryErr := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		body := strings.NewReader(encoded)
		err := WaitAppCreate(ctx, targetInstall.Client, targetInstall.Stack, createParams, body)
		if err != nil {
			if errors.IsConflictError(err.Err) {
				tflog.Info(ctx, "App already installed on target; skipping install and keeping Terraform state in sync",
					map[string]any{"stack": string(targetInstall.Stack), "app": appName})
				return nil
			}
			if strings.Contains(err.Err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %v", err.Err))
			}
			if err.Retryable {
				return resource.RetryableError(fmt.Errorf("retryable error occurred: %v", err.Err))
			}
			return resource.NonRetryableError(err.Err)
		}
		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	return resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		err := WaitAppPoll(ctx, targetInstall.Client, targetInstall.Stack, appName, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "404") {
				return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be created: %s", appName, err))
		}
		return nil
	})
}

func updateApp(ctx context.Context, updateClient v2.ClientInterface, updateStack v2.Stack, appName string, installParams v2.PatchAppVictoriaParams, encoded string, expectedVersion string, verifyVersion bool) error {
	retryErr := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		body := strings.NewReader(encoded)
		err := WaitAppUpdate(ctx, updateClient, updateStack, appName, installParams, body)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			if strings.Contains(err.Error(), "404") {
				return resource.RetryableError(fmt.Errorf("received 404 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error updating app (%s): %s", appName, err))
		}
		return nil
	})
	if retryErr != nil {
		return retryErr
	}

	if !verifyVersion {
		return nil
	}

	return resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		app, err := WaitAppRead(ctx, updateClient, updateStack, appName)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error reading app (%s) after update: %s", appName, err))
		}
		if app.Version == nil || *app.Version != expectedVersion {
			actualVersion := "<nil>"
			if app.Version != nil {
				actualVersion = *app.Version
			}
			return resource.RetryableError(fmt.Errorf("app version (%s) does not match the expected version (%s), retrying", actualVersion, expectedVersion))
		}
		return nil
	})
}

func installOnTarget(ctx context.Context, targetInstall client.TargetFields, appName string, ACSLicensingAck string, encoded string) error {
	localScope := "local"
	return installApp(ctx, targetInstall, appName, ACSLicensingAck, encoded, &localScope)
}

func deleteOnTarget(ctx context.Context, targetInstall client.TargetFields, appName string) error {
	localScope := "local"
	deleteParams := v2.UninstallAppVictoriaParams{
		Scope: &localScope,
	}
	retryErr := resource.RetryContext(ctx, 2*RetryTimeout, func() *resource.RetryError {
		err := WaitAppDelete(ctx, targetInstall.Client, targetInstall.Stack, appName, deleteParams)
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
		err := WaitAppPoll(ctx, targetInstall.Client, targetInstall.Stack, appName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
		}
		return nil
	})
}
