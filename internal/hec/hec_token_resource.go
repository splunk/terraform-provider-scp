package hec

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hashicorp/go-cty/cty"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/status"
	"github.com/splunk/terraform-provider-scp/internal/wait"
)

const (
	ResourceKey          = "scp_hec_tokens"
	NameKey              = "name"
	AllowedIndexesKey    = "allowed_indexes"
	DefaultIndexKey      = "default_index"
	DefaultSourceKey     = "default_source"
	DefaultSourcetypeKey = "default_sourcetype"
	DisabledKey          = "disabled"
	TokenKey             = "token"
	UseAckKey            = "use_ack"
)

func hecTokenResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		NameKey: {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the hec token to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old hec token and recreate with new name).",
		},
		AllowedIndexesKey: {
			Type:     schema.TypeSet,
			Optional: true,
			Computed: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Set of indexes allowed for events with this token",
		},
		DefaultIndexKey: {
			Type:             schema.TypeString,
			Optional:         true,
			Computed:         true,
			Description:      "Index to store generated events",
			ValidateDiagFunc: defaultIndexValidationFunc,
		},
		DefaultSourceKey: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Default source for events with this token",
		},
		DefaultSourcetypeKey: {
			Type:        schema.TypeString,
			Optional:    true,
			Description: "Default sourcetype for events with this token",
		},
		DisabledKey: {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Input disabled indicator: false = Input Not disabled, true = Input disabled",
		},
		TokenKey: {
			ForceNew:    true,
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			Description: "Token value for sending data to collector/event endpoint",
		},
		UseAckKey: {
			Type:        schema.TypeBool,
			Optional:    true,
			Computed:    true,
			Description: "Indexer acknowledgement for this token: false = disabled, true = enabled",
		},
	}
}

func ResourceHecToken() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Hec Token Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageHecTokens " +
			"for more latest, detailed information on attribute requirements and the ACS Hec Token API.",

		CreateContext: resourceHecTokenCreate,
		ReadContext:   resourceHecTokenRead,
		UpdateContext: resourceHecTokenUpdate,
		DeleteContext: resourceHecTokenDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: hecTokenResourceSchema(),
	}
}

func resourceHecTokenCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	hecRequest := parseHecRequest(d)

	createHecRequest := v2.CreateHECJSONRequestBody{
		AllowedIndexes:    hecRequest.AllowedIndexes,
		DefaultIndex:      hecRequest.DefaultIndex,
		DefaultSource:     hecRequest.DefaultSource,
		DefaultSourcetype: hecRequest.DefaultSourcetype,
		Disabled:          hecRequest.Disabled,
		Name:              hecRequest.Name,
		Token:             hecRequest.Token,
		UseAck:            hecRequest.UseAck,
	}

	tflog.Info(ctx, fmt.Sprintf("%+v\n", createHecRequest))

	// Create Hec Token
	err := WaitHecCreate(ctx, acsClient, stack, createHecRequest)
	if err != nil {
		if stateErr := err.(*resource.UnexpectedStateError); stateErr.State == http.StatusText(http.StatusConflict) {
			return diag.Errorf(fmt.Sprintf("Hec (%s) %s", hecRequest.Name, errors.ResourceExistsErr))
		}

		return diag.Errorf(fmt.Sprintf("Error submitting request for hec (%s) to be created: %s", hecRequest.Name, err))
	}

	// Poll Hec until GET returns 200 to confirm hec creation
	err = WaitHecPoll(ctx, acsClient, stack, hecRequest.Name, wait.TargetStatusResourceExists, wait.PendingStatusVerifyCreated)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for hec (%s) to be created: %s", hecRequest.Name, err))
	}

	// Set ID of hec resource to indicate hec has been created
	d.SetId(hecRequest.Name)

	tflog.Info(ctx, fmt.Sprintf("Created hec resource: %s\n", hecRequest.Name))

	// Call readHec to set attributes of hec
	return resourceHecTokenRead(ctx, d, m)

}

func resourceHecTokenRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	hecName := d.Id()

	hec, err := WaitHecRead(ctx, acsClient, stack, hecName)

	if err != nil {
		// if hec not found set id of resource to empty string to remove from state
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), status.HecNotFound) {
			tflog.Info(ctx, fmt.Sprintf("Removing HEC token from state. Not Found error while reading HEC (%s): %s.", hecName, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		} else {
			return diag.Errorf(fmt.Sprintf("Error reading HEC (%s): %s", hecName, err))
		}
	}

	if err := d.Set(NameKey, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(AllowedIndexesKey, hec.AllowedIndexes); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(DefaultIndexKey, hec.DefaultIndex); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(DefaultSourceKey, hec.DefaultSource); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(DefaultSourcetypeKey, hec.DefaultSourcetype); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(DisabledKey, hec.Disabled); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(TokenKey, hec.Token); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(UseAckKey, hec.UseAck); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceHecTokenUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	hecName := d.Id()

	// Retrieve data for each field and create request body
	hecRequest := parseHecRequest(d)
	patchRequest := setPatchRequestBody(d, hecRequest)

	err := WaitHecUpdate(ctx, acsClient, stack, *patchRequest, hecName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error submitting request for hec (%s) to be updated: %s", hecName, err))
	}

	//Poll until fields have been confirmed updated
	err = WaitVerifyHecUpdate(ctx, acsClient, stack, *patchRequest, hecName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for hec (%s) to be updated: %s", hecName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("updated hec resource: %s\n", hecName))
	return resourceHecTokenRead(ctx, d, m)
}

func resourceHecTokenDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	hecName := d.Id()

	err := WaitHecDelete(ctx, acsClient, stack, hecName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error deleting hec (%s): %s", hecName, err))
	}

	//Poll hec until GET returns 404 Not found - hec has been deleted
	err = WaitHecPoll(ctx, acsClient, stack, hecName, wait.TargetStatusResourceDeleted, wait.PendingStatusVerifyDeleted)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for hec (%s) to be deleted: %s", hecName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("deleted hec resource: %s\n", hecName))
	return nil
}

// Parse and return hec attribute data set in configuration
func parseHecRequest(d *schema.ResourceData) *v2.HecSpec {
	hecRequest := v2.HecSpec{}

	hecRequest.Name = d.Get(NameKey).(string)

	if allowedIndexes, _ := d.GetOk(AllowedIndexesKey); allowedIndexes != nil {
		allowedIndexesSet := allowedIndexes.(*schema.Set)
		parsedData := make([]string, 0)
		for _, allowedIndex := range allowedIndexesSet.List() {
			parsedData = append(parsedData, allowedIndex.(string))
		}
		hecRequest.AllowedIndexes = &parsedData
	}

	if defaultIndex, _ := d.GetOk(DefaultIndexKey); defaultIndex != nil {
		parsedData := defaultIndex.(string)
		hecRequest.DefaultIndex = &parsedData
	}

	if defaultSource, _ := d.GetOk(DefaultSourceKey); defaultSource != nil {
		parsedData := defaultSource.(string)
		hecRequest.DefaultSource = &parsedData
	}

	if defaultSourcetype, _ := d.GetOk(DefaultSourcetypeKey); defaultSourcetype != nil {
		parsedData := defaultSourcetype.(string)
		hecRequest.DefaultSourcetype = &parsedData
	}

	if disabled, _ := d.GetOk(DisabledKey); disabled != nil {
		parsedData := disabled.(bool)
		hecRequest.Disabled = &parsedData
	}

	if token, _ := d.GetOk(TokenKey); token != nil {
		parsedData := token.(string)
		hecRequest.Token = &parsedData
	}

	if useAck, _ := d.GetOk(UseAckKey); useAck != nil {
		parsedData := useAck.(bool)
		hecRequest.UseAck = &parsedData
	}
	return &hecRequest
}

// Only set params in PatchHec Request if given key has been changed in configuration
func setPatchRequestBody(d *schema.ResourceData, hecRequest *v2.HecSpec) *v2.PatchHECJSONRequestBody {
	patchRequest := v2.PatchHECJSONRequestBody{}

	if d.HasChange(AllowedIndexesKey) {
		patchRequest.AllowedIndexes = hecRequest.AllowedIndexes
	}

	if d.HasChange(DefaultIndexKey) {
		patchRequest.DefaultIndex = hecRequest.DefaultIndex
	}

	if d.HasChange(TokenKey) {
		patchRequest.Token = hecRequest.Token
	}

	if d.HasChange(DefaultSourceKey) {
		patchRequest.DefaultSource = hecRequest.DefaultSource
	}

	if d.HasChange(DefaultSourcetypeKey) {
		patchRequest.DefaultSourcetype = hecRequest.DefaultSourcetype
	}

	if d.HasChange(DisabledKey) {
		patchRequest.Disabled = hecRequest.Disabled
	}

	if d.HasChange(UseAckKey) {
		patchRequest.UseAck = hecRequest.UseAck
	}
	return &patchRequest
}

func defaultIndexValidationFunc(v interface{}, p cty.Path) diag.Diagnostics {
	var diags diag.Diagnostics
	currentValue := v.(string)

	// Splunkd automatically populates the default index (based on allowed indexes field)if not default index is provided.
	// If the user omits this key, the value is computed and stored in the state
	// However, we do not want users to explicitly hard code the value to be empty string
	if strings.TrimSpace(currentValue) == "" {
		errorDiag := diag.Diagnostic{
			Severity: diag.Error,
			Summary:  "invalid value",
			Detail:   fmt.Sprintf("%s cannot be an empty string. Either omit it or pick an index from %s", DefaultIndexKey, AllowedIndexesKey),
		}
		diags = append(diags, errorDiag)
	}
	return diags
}
