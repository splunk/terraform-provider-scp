package users

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	"strings"
)

const (
	ResourceKey = "scp_users"

	schemaKeyName                     = "name"
	schemaKeyPassword                 = "password"
	schemaKeyOldPassword              = "old_password"
	schemaKeyDefaultApp               = "default_app"
	schemaKeyEmail                    = "email"
	schemaKeyForceChangePass          = "force_change_pass"
	schemaKeyFullName                 = "full_name"
	schemaKeyRoles                    = "roles"
	schemaKeyFederatedSearchManageAck = "federated_search_manage_ack"
	schemaKeyDefaultAppSource         = "default_app_source"
	schemaKeyLastSuccessfulLogin      = "last_successful_login"
	schemaKeyLockedOut                = "locked_out"
)

// UserInfo represents the generic user data model
type UserInfo struct {
	DefaultApp      *string   `json:"defaultApp,omitempty"`
	Email           *string   `json:"email,omitempty"`
	ForceChangePass *bool     `json:"forceChangePass,omitempty"`
	FullName        *string   `json:"fullName,omitempty"`
	OldPassword     *string   `json:"oldPassword,omitempty"`
	Password        *string   `json:"password,omitempty"`
	Roles           *[]string `json:"roles,omitempty"`
	Name            string    `json:"name"`
}

func userResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		schemaKeyName: {
			Type:     schema.TypeString,
			Required: true,
			ForceNew: true,
			Description: "The name of the user to create. Can not be updated after creation, " +
				"if changed in config file terraform will propose a replacement (delete old user and recreate with new name).",
		},
		schemaKeyPassword: {
			Type:     schema.TypeString,
			Required: true,
			Description: "The password of the user to create, or the new password to update. " +
				"To protect your password, you can replace the credentials with variables configured with the sensitive flag, " +
				"and set values for these variables using environment variables or with a .tfvars file. " +
				"Please refer to https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables for more details.",
		},
		schemaKeyOldPassword: {
			Type:     schema.TypeString,
			Optional: true,
			Required: false,
			Description: "The old password of the user to update. " +
				"To protect your password, you can replace the credentials with variables configured with the sensitive flag, " +
				"and set values for these variables using environment variables or with a .tfvars file. " +
				"Please refer to https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables for more details.",
		},
		schemaKeyDefaultApp: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Set the default app for this user. Setting this here overrides the default app inherited from the user's role(s).",
		},
		schemaKeyEmail: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The email of the user to create.",
		},
		schemaKeyForceChangePass: {
			Type:        schema.TypeBool,
			Optional:    true,
			Default:     true,
			Description: "To force a change of password on the user's first login, set forceChangePass to \"true\".",
		},
		schemaKeyFullName: {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "The full name of the user to create.",
		},
		schemaKeyRoles: {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Assign one of more roles to this user. The user will inherit all the settings and capabilities from those roles.",
		},
		schemaKeyFederatedSearchManageAck: {
			Type:     schema.TypeString,
			Optional: true,
			Description: "If any role contains the 'fsh_manage' capability you must set this attribute to a value of \"Y\". " +
				"This header acknowledges that a role with the fsh_manage capability can send search results outside the compliant environment.",
		},
		schemaKeyDefaultAppSource: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "Default app source of the user.",
		},
		schemaKeyLastSuccessfulLogin: {
			Type:        schema.TypeString,
			Computed:    true,
			Description: "(Read-only) Last successful login timestamp of the user.",
		},
		schemaKeyLockedOut: {
			Type:        schema.TypeBool,
			Computed:    true,
			Description: "(Read-only) Whether the user account has been locked out.",
		},
	}
}

func ResourceUser() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "User Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageRoles " +
			"for more latest, detailed information on attribute requirements and the ACS Users API.",

		CreateContext: resourceUserCreate,
		ReadContext:   resourceUserRead,
		UpdateContext: resourceUserUpdate,
		DeleteContext: resourceUserDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: userResourceSchema(),
	}
}

func resourceUserCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	userInfo := parseUserRequest(d)
	// We won't allow customers to set its value and the default value will always be false,
	// so that customers will need to explicitly declare a role for terraform to manage.
	createRole := false
	createRequest := v2.CreateUserJSONRequestBody{
		CreateRole:      &createRole,
		DefaultApp:      userInfo.DefaultApp,
		Email:           userInfo.Email,
		ForceChangePass: userInfo.ForceChangePass,
		FullName:        userInfo.FullName,
		Name:            userInfo.Name,
		Password:        *userInfo.Password,
		Roles:           userInfo.Roles,
	}
	userParam := parseUserParams(d)
	createParam := v2.CreateUserParams{
		FederatedSearchManageAck: userParam,
	}

	tflog.Info(ctx, fmt.Sprintf("%+v\n", createRequest))

	err := WaitUserCreate(ctx, acsClient, stack, createParam, createRequest)
	if err != nil {
		if errors.IsConflictError(err) {
			return diag.Errorf(fmt.Sprintf("User (%s) already exists, use a different name to create user or use terraform import to bring current user under terraform management", createRequest.Name))
		}

		return diag.Errorf(fmt.Sprintf("Error submitting request for user (%s) to be created: %s", createRequest.Name, err))
	}

	// Set ID of user resource to indicate user has been created
	d.SetId(createRequest.Name)

	tflog.Info(ctx, fmt.Sprintf("Created user resource: %s\n", createRequest.Name))

	// Call readUser to set attributes of user
	return resourceUserRead(ctx, d, m)
}

func resourceUserRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	tflog.Info(ctx, fmt.Sprintf("resourceUserRead invoked"))
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	userName := d.Id()

	user, err := WaitUserRead(ctx, acsClient, stack, userName)

	if err != nil {
		// if user not found set id of resource to empty string to remove from state
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), status.ErrUserNotFound) {
			tflog.Info(ctx, fmt.Sprintf("Removing user from state. Not Found error while reading user (%s): %s.", userName, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		} else {
			return diag.Errorf(fmt.Sprintf("Error reading user (%s): %s", userName, err))
		}
	}

	if err := d.Set(schemaKeyName, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyDefaultApp, user.DefaultApp); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyDefaultAppSource, user.DefaultAppSource); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyEmail, user.Email); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyFullName, user.FullName); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyLastSuccessfulLogin, user.LastSuccessfulLogin); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyLockedOut, user.LockedOut); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeyRoles, user.Roles); err != nil {
		return diag.FromErr(err)
	}
	return nil
}

func resourceUserUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	userName := d.Id()

	// Retrieve data for each field and create request body
	userInfo := parseUserRequest(d)
	patchRequest := v2.PatchUserJSONRequestBody{
		DefaultApp:      userInfo.DefaultApp,
		Email:           userInfo.Email,
		ForceChangePass: userInfo.ForceChangePass,
		FullName:        userInfo.FullName,
		OldPassword:     userInfo.OldPassword,
		Password:        userInfo.Password,
		Roles:           userInfo.Roles,
	}
	userParam := parseUserParams(d)
	patchParam := v2.PatchUserParams{
		FederatedSearchManageAck: userParam,
	}

	err := WaitUserUpdate(ctx, acsClient, stack, patchParam, patchRequest, userName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error submitting request for user (%s) to be updated: %s", userName, err))
	}

	//Poll until fields have been confirmed updated, good to keep even though resource is sync
	err = WaitVerifyUserUpdate(ctx, acsClient, stack, patchRequest, userName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for user (%s) to be updated: %s", userName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("updated hec resource: %s\n", userName))
	return resourceUserRead(ctx, d, m)
}

func resourceUserDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	userName := d.Id()

	err := WaitUserDelete(ctx, acsClient, stack, userName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error deleting user (%s): %s", userName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("deleted user resource: %s\n", userName))
	return nil
}

func parseUserRequest(d *schema.ResourceData) *UserInfo {
	userInfo := UserInfo{}

	userInfo.Name = d.Get(schemaKeyName).(string)

	if password, ok := d.GetOk(schemaKeyPassword); ok {
		parsedData := password.(string)
		userInfo.Password = &parsedData
	}

	if oldPassword, ok := d.GetOk(schemaKeyOldPassword); ok {
		parsedData := oldPassword.(string)
		userInfo.OldPassword = &parsedData
	}

	if defaultApp, ok := d.GetOk(schemaKeyDefaultApp); ok {
		parsedData := defaultApp.(string)
		userInfo.DefaultApp = &parsedData
	}

	if email, ok := d.GetOk(schemaKeyEmail); ok {
		parsedData := email.(string)
		userInfo.Email = &parsedData
	}

	if forceChangePass, ok := d.GetOk(schemaKeyForceChangePass); ok {
		parsedData := forceChangePass.(bool)
		userInfo.ForceChangePass = &parsedData
	}

	if fullName, ok := d.GetOk(schemaKeyFullName); ok {
		parsedData := fullName.(string)
		userInfo.FullName = &parsedData
	}

	if roles, ok := d.GetOk(schemaKeyRoles); ok {
		parsedData := utils.ParseSetValues(roles)
		userInfo.Roles = &parsedData
	}

	return &userInfo
}

func parseUserParams(d *schema.ResourceData) *v2.FederatedSearchManage {
	if fshManaged, ok := d.GetOk(schemaKeyFederatedSearchManageAck); ok {
		parsedData := fshManaged.(string)
		return (*v2.FederatedSearchManage)(&parsedData)
	}

	return nil
}
