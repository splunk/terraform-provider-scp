package roles

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/utils"
)

const (
	ResourceKey = "scp_roles"

	schemaKeyName                      = "name"
	schemaKeyCapabilities              = "capabilities"
	schemaKeyCumulativeRTSrchJobsQuota = "cumulative_rt_srch_jobs_quota"
	schemaKeyCumulativeSrchJobsQuota   = "cumulative_srch_jobs_quota"
	schemaKeyDefaultApp                = "default_app"
	schemaKeyImportedRoles             = "imported_roles"
	schemaKeyRTSrchJobsQuota           = "rt_srch_jobs_quota"
	schemaKeySrchJobsQuota             = "srch_jobs_quota"
	schemaKeySrchDiskQuota             = "srch_disk_quota"
	schemaKeySrchFilter                = "srch_filter"
	schemaKeySrchIndexesAllowed        = "srch_indexes_allowed"
	schemaKeySrchIndexesDefault        = "srch_indexes_default"
	schemaKeySrchTimeEarliest          = "srch_time_earliest"
	schemaKeySrchTimeWin               = "srch_time_win"
	schemaKeyFederatedSearchManageAck  = "federated_search_manage_ack"
)

func roleResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		schemaKeyName: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: "The name of the role to create. Can not be updated after creation, " +
				"if changed in config file terraform will propose a replacement (delete old role and recreate with new name).",
		},
		schemaKeyCapabilities: {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "The capabilities attached to the role.",
		},
		schemaKeyCumulativeRTSrchJobsQuota: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum number of concurrently running real-time searches that all members of this role can have. " +
				"The value must be a non-negative number.",
		},
		schemaKeyCumulativeSrchJobsQuota: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum number of concurrently running historical searches that all members of this role can have. " +
				"The value must be a non-negative number.",
		},
		schemaKeyDefaultApp: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Set the default app for this role.",
		},
		schemaKeyImportedRoles: {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of other roles and their associated capabilities that should be imported.",
		},
		schemaKeyRTSrchJobsQuota: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum number of concurrently running real-time searches a member of this role can have. " +
				"The value must be a non-negative number.",
		},
		schemaKeySrchJobsQuota: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum number of concurrently running historical searches a member of this role can have. " +
				"The value must be a non-negative number.",
		},
		schemaKeySrchDiskQuota: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum amount of disk space (MB) that can be used by search jobs of a user that belongs to this role. " +
				"The value must be a non-negative number.",
		},
		schemaKeySrchFilter: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "List of search filters for this Role.",
		},
		schemaKeySrchIndexesAllowed: {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of indexes this role is allowed to search.",
		},
		schemaKeySrchIndexesDefault: {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "List of indexes to search when no index is specified.",
		},
		schemaKeySrchTimeEarliest: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum amount of time that searches of users from this role will be allowed to run. " +
				"A value of -1 means unset, 0 means infinite. Any other value is the amount of time in seconds, for example, 300 would mean 300s.",
		},
		schemaKeySrchTimeWin: {
			Type:     schema.TypeInt,
			Optional: true,
			Computed: true,
			Description: "Maximum time span of a search, in seconds. " +
				"A value of -1 means unset, 0 means infinite. Any other value is the amount of time in seconds, for example, 300 would mean 300s.",
		},

		schemaKeyFederatedSearchManageAck: {
			Type:     schema.TypeString,
			Optional: true,
			Description: "If 'imported_roles' or 'capabilities' contains the 'fsh_manage' capability, you must set this attribute to a value of \"Y\". " +
				"This header acknowledges that a role with the 'fsh_manage' capability can send search results outside the compliant environment.",
		},
	}
}

func ResourceRole() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Role Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageRoles " +
			"for more latest, detailed information on attribute requirements and the ACS Roles API.",

		CreateContext: resourceRoleCreate,
		ReadContext:   resourceRoleRead,
		UpdateContext: resourceRoleUpdate,
		DeleteContext: resourceRoleDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: roleResourceSchema(),
	}
}

func resourceRoleCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	roleRequest, roleName := parseRoleRequest(d)
	createRequest := v2.CreateRoleJSONRequestBody{
		Name:         roleName,
		RolesRequest: *roleRequest,
	}
	roleParam := parseRoleParams(d)
	createParam := v2.CreateRoleParams{
		FederatedSearchManageAck: roleParam,
	}

	tflog.Info(ctx, fmt.Sprintf("%+v\n", createRequest))

	err := WaitRoleCreate(ctx, acsClient, stack, createParam, createRequest)
	if err != nil {
		if errors.IsConflictError(err) {
			return diag.Errorf("Role (%s) already exists, use a different name to create role or use terraform import to bring current role under terraform management", createRequest.Name)
		}

		return diag.Errorf("Error submitting request for role (%s) to be created: %s", createRequest.Name, err)
	}

	// Set ID of role resource to indicate role has been created
	d.SetId(createRequest.Name)

	tflog.Info(ctx, fmt.Sprintf("Created role resource: %s\n", createRequest.Name))

	// Call readRole to set attributes of role
	return resourceRoleRead(ctx, d, m)
}

func resourceRoleRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, "resourceRoleRead invoked")
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	roleName := d.Id()

	roleResponse, err := WaitRoleRead(ctx, acsClient, stack, roleName)

	if err != nil {
		// if role not found set id of resource to empty string to remove from state
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), status.ErrRoleNotFound) {
			tflog.Info(ctx, fmt.Sprintf("Removing role from state. Not Found error while reading role (%s): %s.", roleName, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		}
		return diag.Errorf("Error reading role (%s): %s", roleName, err)
	}

	if err := d.Set(schemaKeyName, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyCapabilities, roleResponse.Capabilities); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyCumulativeRTSrchJobsQuota, roleResponse.CumulativeRTSrchJobsQuota); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyCumulativeSrchJobsQuota, roleResponse.CumulativeSrchJobsQuota); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyDefaultApp, roleResponse.DefaultApp); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyImportedRoles, roleResponse.Imported.Roles); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyRTSrchJobsQuota, roleResponse.RtSrchJobsQuota); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchJobsQuota, roleResponse.SrchJobsQuota); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchDiskQuota, roleResponse.SrchDiskQuota); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchFilter, roleResponse.SrchFilter); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchIndexesAllowed, roleResponse.SrchIndexesAllowed); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchIndexesDefault, roleResponse.SrchIndexesDefault); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchTimeEarliest, roleResponse.SrchTimeEarliest); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySrchTimeWin, roleResponse.SrchTimeWin); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceRoleUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	patchRequest, roleName := parseRoleRequest(d)

	roleParam := parseRoleParams(d)
	patchParam := v2.PatchRoleInfoParams{
		FederatedSearchManageAck: roleParam,
	}
	tflog.Info(ctx, fmt.Sprintf("updated role resource: %d\n", patchRequest.CumulativeRTSrchJobsQuota))

	patchRequestBody := v2.PatchRoleInfoJSONRequestBody{
		RolesInfo:                 patchRequest.RolesInfo,
		CumulativeRTSrchJobsQuota: patchRequest.CumulativeRTSrchJobsQuota,
		CumulativeSrchJobsQuota:   patchRequest.CumulativeSrchJobsQuota,
		DefaultApp:                patchRequest.DefaultApp,
		ImportedRoles:             patchRequest.ImportedRoles,
	}
	err := WaitRoleUpdate(ctx, acsClient, stack, patchParam, patchRequestBody, roleName)
	if err != nil {
		return diag.Errorf("Error submitting request for role (%s) to be updated: %s", roleName, err)
	}

	//Poll until fields have been confirmed updated, good to keep even though resource is sync
	err = WaitVerifyRoleUpdate(ctx, acsClient, stack, patchRequestBody, roleName)
	if err != nil {
		return diag.Errorf("%s", fmt.Sprintf("Error waiting for role (%s) to be updated: %s", roleName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("updated role resource: %s\n", roleName))
	return resourceRoleRead(ctx, d, m)
}

func resourceRoleDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	roleName := d.Id()

	err := WaitRoleDelete(ctx, acsClient, stack, roleName)
	if err != nil {
		return diag.Errorf("%s", fmt.Sprintf("Error deleting role (%s): %s", roleName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("deleted role resource: %s\n", roleName))
	return nil
}

func parseRoleRequest(d *schema.ResourceData) (*v2.RolesRequest, string) {
	rolesRequest := v2.RolesRequest{}
	rolesInfo := v2.RolesInfo{}
	rolesRequest.RolesInfo = rolesInfo

	name := d.Get(schemaKeyName).(string)

	// RolesRequest attributes

	//	workaround to allow 0 value, repeated for all fields where 0 is valid value
	if value, ok := d.GetOk(schemaKeyCumulativeRTSrchJobsQuota); ok {
		parsedData := value.(int)
		rolesRequest.CumulativeRTSrchJobsQuota = &parsedData
	} else if d.HasChange(schemaKeyCumulativeRTSrchJobsQuota) {
		_, newVal := d.GetChange(schemaKeyCumulativeRTSrchJobsQuota)
		parsedData := newVal.(int)
		rolesRequest.CumulativeRTSrchJobsQuota = &parsedData
	}

	if value, ok := d.GetOk(schemaKeyCumulativeSrchJobsQuota); ok {
		parsedData := value.(int)
		rolesRequest.CumulativeSrchJobsQuota = &parsedData
	} else if d.HasChange(schemaKeyCumulativeSrchJobsQuota) {
		_, newVal := d.GetChange(schemaKeyCumulativeSrchJobsQuota)
		parsedData := newVal.(int)
		rolesRequest.CumulativeSrchJobsQuota = &parsedData
	}

	if value, ok := d.GetOk(schemaKeyDefaultApp); ok {
		parsedData := value.(string)
		rolesRequest.DefaultApp = &parsedData
	}

	if values, ok := d.GetOk(schemaKeyImportedRoles); ok {
		parsedData := utils.ParseSetValues(values)
		rolesRequest.ImportedRoles = &parsedData
	}

	// RolesInfo attributes
	if values, ok := d.GetOk(schemaKeyCapabilities); ok {
		parsedData := utils.ParseSetValues(values)
		rolesRequest.Capabilities = &parsedData
	}

	// workaround to allow zero value
	if value, ok := d.GetOk(schemaKeyRTSrchJobsQuota); ok {
		parsedData := value.(int)
		rolesRequest.RtSrchJobsQuota = &parsedData
	} else if d.HasChange(schemaKeyRTSrchJobsQuota) {
		_, newVal := d.GetChange(schemaKeyRTSrchJobsQuota)
		parsedData := newVal.(int)
		rolesRequest.RtSrchJobsQuota = &parsedData
	}

	if value, ok := d.GetOk(schemaKeySrchDiskQuota); ok {
		parsedData := value.(int)
		rolesRequest.SrchDiskQuota = &parsedData
	} else if d.HasChange(schemaKeySrchDiskQuota) {
		_, newVal := d.GetChange(schemaKeySrchDiskQuota)
		parsedData := newVal.(int)
		rolesRequest.SrchDiskQuota = &parsedData
	}

	if value, ok := d.GetOk(schemaKeySrchFilter); ok {
		parsedData := value.(string)
		rolesRequest.SrchFilter = &parsedData
	}

	if values, ok := d.GetOk(schemaKeySrchIndexesAllowed); ok {
		parsedData := utils.ParseSetValues(values)
		rolesRequest.SrchIndexesAllowed = &parsedData
	}

	if values, ok := d.GetOk(schemaKeySrchIndexesDefault); ok {
		parsedData := utils.ParseSetValues(values)
		rolesRequest.SrchIndexesDefault = &parsedData
	}

	if value, ok := d.GetOk(schemaKeySrchJobsQuota); ok {
		parsedData := value.(int)
		rolesRequest.SrchJobsQuota = &parsedData
	} else if d.HasChange(schemaKeySrchJobsQuota) {
		_, newVal := d.GetChange(schemaKeySrchJobsQuota)
		parsedData := newVal.(int)
		rolesRequest.SrchJobsQuota = &parsedData
	}

	if value, ok := d.GetOk(schemaKeySrchTimeEarliest); ok {
		parsedData := value.(int)
		rolesRequest.SrchTimeEarliest = &parsedData
	} else if d.HasChange(schemaKeySrchTimeEarliest) {
		_, newVal := d.GetChange(schemaKeySrchTimeEarliest)
		parsedData := newVal.(int)
		rolesRequest.SrchTimeEarliest = &parsedData
	}

	if value, ok := d.GetOk(schemaKeySrchTimeWin); ok {
		parsedData := value.(int)
		rolesRequest.SrchTimeWin = &parsedData
	} else if d.HasChange(schemaKeySrchTimeWin) {
		_, newVal := d.GetChange(schemaKeySrchTimeWin)
		parsedData := newVal.(int)
		rolesRequest.SrchTimeWin = &parsedData
	}
	return &rolesRequest, name
}

func parseRoleParams(d *schema.ResourceData) *v2.FederatedSearchManage {
	if fshManaged, ok := d.GetOk(schemaKeyFederatedSearchManageAck); ok {
		parsedData := fshManaged.(string)
		return (*v2.FederatedSearchManage)(&parsedData)
	}

	return nil
}
