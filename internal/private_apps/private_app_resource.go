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
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

const (
	ResourceKey  = "scp_private_app"
	AcsLegalAck  = "acs_legal_ack"
	RetryTimeout = 20 * time.Minute
)

func privateAppSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the private app",
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
			Required:    true,
			Description: "Whether the private app has been pre-vetted using AppInspect.",
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
	tflog.Info(ctx, "[BETA] Private Apps: This feature is in beta release.")
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack
	splunkbase := false

	// Acquire lock for app operations on this stack
	lockManager := locks.GetAppLockManager()
	unlock := lockManager.LockAppOperation(ctx, stack, "create")
	defer unlock()

	validationStatus := resourceData.Get("pre_vetted").(bool)
	if !validationStatus {
		return diag.Errorf("App must be pre-vetted before it can be installed. Set 'pre_vetted' to true manually after vetting the app or use scp_app_validation data source to automate the process.")
	}

	ACSLicensingAck := resourceData.Get("acs_legal_ack").(string)

	installParams := v2.InstallAppVictoriaParams{
		Splunkbase:  &splunkbase,
		ACSLegalAck: &ACSLicensingAck,
	}

	var appID string
	retriesOfBadPackage := 0

	err := resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		fileData, errFileRead := os.ReadFile(resourceData.Get("filename").(string))
		if errFileRead != nil {
			return resource.NonRetryableError(errFileRead)
		}
		body := bytes.NewReader(fileData)
		output, err := WaitAppCreate(ctx, acsClient, stack, installParams, body)
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
		err := WaitAppPoll(ctx, acsClient, stack, appID, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
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

	resourceData.SetId(appID)
	return nil
}

func resourcePrivateAppRead(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	acsProvider := m.(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	appName := resourceData.Id()
	_, err := WaitAppRead(ctx, acsClient, stack, appName)

	if err != nil {
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), "404-app-not-found") {
			return nil
		}
		return diag.Errorf("Error reading app (%s): %s", appName, err)
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

	appName := resourceData.Id()

	err := resource.RetryContext(ctx, 2*RetryTimeout, func() *resource.RetryError {
		err := WaitAppDelete(ctx, acsClient, stack, appName)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(err)
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("Error deleting app (%s): %s", appName, err)
	}

	err = resource.RetryContext(ctx, RetryTimeout, func() *resource.RetryError {
		err := WaitAppPoll(ctx, acsClient, stack, appName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
		if err != nil {
			if strings.Contains(err.Error(), "503") {
				return resource.RetryableError(fmt.Errorf("received 503 error, retrying: %w", err))
			}
			return resource.NonRetryableError(fmt.Errorf("error waiting for app (%s) to be deleted: %s", appName, err))
		}
		return nil
	})
	if err != nil {
		return diag.Errorf("Error waiting for app (%s) to be deleted: %s", appName, err)
	}

	return nil
}

func resourcePrivateAppUpdate(ctx context.Context, resourceData *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "[BETA] Private Apps: This feature is in beta release.")

	err := resourcePrivateAppDelete(ctx, resourceData, m)
	if err.HasError() {
		return err
	}

	appName := resourceData.Get("name").(string)
	tflog.Info(ctx, fmt.Sprintf("Recreating app (%s) after deletion.", appName))

	err = resourcePrivateAppCreate(ctx, resourceData, m)
	if err.HasError() {
		return diag.Errorf("Error recreating app (%s) after deletion: %v", appName, err)
	}
	return nil
}
