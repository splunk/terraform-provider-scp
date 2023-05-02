package ipallowlists

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
)

type ipAllowlistRequestBody struct {
	feature string
	subnets []string
}

const (
	ResourceKey = "scp_ip_allowlists"
)

func ipAllowlistResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"feature": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "Feature is a specified component in your Splunk Cloud Platform. Eg: search-api, hec, etc. ",
		},
		"subnets": {
			Type:     schema.TypeList,
			Required: true,
			ForceNew: false,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Subnets is a list of IP addresses that have access to the corresponding feature.",
		},
	}
}

func ResourceIPAllowlist() *schema.Resource {
	return &schema.Resource{
		Description: "IP Allowlist Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ConfigureIPAllowList " +
			"for more latest, detailed information on attribute requirements and the ACS IP Allowlist API.",

		CreateContext: resourceIPAllowlistCreate,
		ReadContext:   resourceIPAllowlistRead,
		UpdateContext: resourceIPAllowlistUpdate,
		DeleteContext: resourceIPAllowlistDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: ipAllowlistResourceSchema(),
	}
}

func resourceIPAllowlistCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	request := parseIPAllowlistRequest(d)

	tflog.Info(ctx, fmt.Sprintf("%+v\n", request.subnets))

	// Add new subnets
	err := WaitIPAllowlistCreate(ctx, acsClient, stack, v2.Feature(request.feature), request.subnets)
	if err != nil {
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), "unknown access feature") {
			tflog.Info(ctx, fmt.Sprintf("Invalid IP Allowlist feature (%s): %s.", request.feature, err))
		}
		return diag.Errorf(fmt.Sprintf("Error submitting request for IP allowlist (%s) to be created: %s", request.feature, err))
	}

	// Set ID of index resource to indicate index has been created
	d.SetId(request.feature)

	tflog.Info(ctx, fmt.Sprintf("Created IP Allowlist resource for feature: %s\n", request.feature))

	// Call readIndex to set attributes of index
	return resourceIPAllowlistRead(ctx, d, m)
}

func resourceIPAllowlistRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	feature := d.Id()

	subnets, err := WaitIPAllowlistRead(ctx, acsClient, stack, feature)

	if err != nil {
		// if feature not found set id of resource to empty string to remove from state
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), "unknown access feature") {
			tflog.Info(ctx, fmt.Sprintf("Invalid IP Allowlist feature (%s): %s.", feature, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		} else {
			return diag.Errorf(fmt.Sprintf("Error reading ip allowlist (%s): %s", feature, err))
		}
	}

	if err := d.Set("feature", d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("subnets", subnets); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceIPAllowlistUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("method not implemented"))
}

func resourceIPAllowlistDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	return diag.FromErr(fmt.Errorf("method not implemented"))
}

func parseIPAllowlistRequest(d *schema.ResourceData) *ipAllowlistRequestBody {
	request := ipAllowlistRequestBody{}

	request.feature = d.Get("feature").(string)

	rawSubnets := d.Get("subnets").([]interface{})
	subnets := make([]string, len(rawSubnets))
	for i, v := range rawSubnets {
		subnets[i] = v.(string)
	}
	request.subnets = subnets

	return &request
}
