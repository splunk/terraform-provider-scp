package appvalidation

import (
	"context"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/splunk/terraform-provider-scp/client"
)

const (
	DataSourceKey                          = "scp_app_validation"
	missingAppValidationCredentialsMessage = "splunk_username and splunk_password must be configured in the provider to use app validation. Configure provider attributes splunk_username and splunk_password, or set SPLUNK_USERNAME and SPLUNK_PASSWORD, so the provider can create the AppInspect login token required by the AppInspect API."
)

func privateAppValidationSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"request_id": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The ID of the validation request.",
		},
		"status": {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "The current status of the validation request.",
		},
		"pre_vetted": {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "Whether the app has been pre-vetted successfully.",
		},
	}
}

func DataSourcePrivateAppValidation() *schema.Resource {
	return &schema.Resource{
		Description: "Use this data source to check the validation status of a private app." +
			"Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageApps for more latest, detailed information on attribute requirements and the ACS API.",
		ReadContext: dataSourcePrivateAppValidationRead,
		Schema:      privateAppValidationSchema(),
	}
}

func dataSourcePrivateAppValidationRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	acsProvider := m.(*client.ACSProvider)
	if acsProvider.AppInspectClient == nil {
		return diag.Errorf(missingAppValidationCredentialsMessage)
	}
	appInspectClient := *acsProvider.AppInspectClient

	requestID := d.Get("request_id").(string)

	validation, err := WaitAppValidationRead(ctx, appInspectClient, requestID)
	if err != nil {
		return diag.FromErr(err)
	}

	if err = d.Set("request_id", validation.RequestID); err != nil {
		return diag.FromErr(err)
	}
	if err = d.Set("status", validation.Status); err != nil {
		return diag.FromErr(err)
	}

	isValidated := strings.ToLower(validation.Status) == "success" && validation.Info.Error == 0 && validation.Info.Failure == 0
	if err = d.Set("pre_vetted", isValidated); err != nil {
		return diag.FromErr(err)
	}

	if !isValidated {
		return diag.Errorf("App validation failed: status=%s, errors=%d, failures=%d", strings.ToLower(validation.Status), validation.Info.Error, validation.Info.Failure)
	}

	d.SetId(validation.RequestID)

	return nil
}
