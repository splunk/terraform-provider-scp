package ipv6allowlists

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/utils"
)

const (
	ResourceKey = "scp_ip_v6_allowlists"

	schemaKeyFeature = "feature"
	schemaKeySubnets = "subnets"
)

func ipv6AllowlistResourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		schemaKeyFeature: {
			Type:     schema.TypeString,
			Required: true,
			Description: "Feature is a specified component in your Splunk Cloud Platform. Eg: search-api, hec, etc. No two " +
				"resources should have the same feature. Use this value as the resource name itself to enforce this rule. ",
		},
		schemaKeySubnets: {
			Type:     schema.TypeSet,
			Required: true,
			Elem: &schema.Schema{
				Type: schema.TypeString,
			},
			Description: "Subnets is a list of IPv6 addresses that have access to the corresponding feature.",
			MinItems:    1,
		},
	}
}

func ResourceIPv6Allowlist() *schema.Resource {
	return &schema.Resource{
		Description: "IPv6 Allowlist Resource. Please see documentation to understand unique behavior regarding naming and delete operation. " +
			"Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ConfigureIPAllowList " +
			"for more latest, detailed information on attribute requirements and the ACS IP Allowlist API.",

		CreateContext: resourceIPv6AllowlistCreate,
		ReadContext:   resourceIPv6AllowlistRead,
		UpdateContext: resourceIPv6AllowlistUpdate,
		DeleteContext: resourceIPv6AllowlistDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: ipv6AllowlistResourceSchema(),
	}
}

func resourceIPv6AllowlistCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	feature, _, newSubnetsSet := parseIPv6AllowlistRequest(d)
	addSubnets := utils.GetSubnetsFromSet(newSubnetsSet)

	// Add new subnets
	err := WaitIPv6AllowlistCreate(ctx, acsClient, stack, v2.Feature(feature), addSubnets)
	if err != nil {
		if errors.IsUnknownFeatureError(err) {
			tflog.Info(ctx, fmt.Sprintf("Invalid IPv6 Allowlist feature (%s): %s.", feature, err))
		}
		return diag.Errorf("Error submitting request for IPv6 allowlist (%s) to be created: %s", feature, err)
	}

	// Set ID of index resource to indicate index has been created
	d.SetId(feature)
	tflog.Info(ctx, fmt.Sprintf("Created IPv6 Allowlist resource for feature: %s\n", feature))

	// Call readIndex to set attributes of index
	return resourceIPv6AllowlistRead(ctx, d, m)
}

func resourceIPv6AllowlistRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	feature := d.Id()

	subnets, err := WaitIPv6AllowlistRead(ctx, acsClient, stack, feature)

	if err != nil {
		// if feature not found set id of resource to empty string to remove from state
		if errors.IsUnknownFeatureError(err) {
			tflog.Info(ctx, fmt.Sprintf("Invalid IPv6 Allowlist feature (%s): %s.", feature, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		}
		return diag.Errorf("Error reading ipv6 allowlist (%s): %s", feature, err)
	}

	if err := d.Set(schemaKeyFeature, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySubnets, subnets); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceIPv6AllowlistUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Determine the changes to the subnets for a feature
	feature, oldSubnetsSet, newSubnetsSet := parseIPv6AllowlistRequest(d)
	addSubnets := utils.GetSubnetsFromSet(newSubnetsSet.Difference(oldSubnetsSet))
	deleteSubnets := utils.GetSubnetsFromSet(oldSubnetsSet.Difference(newSubnetsSet))

	// do not allow feature name to be changed
	if d.HasChange(schemaKeyFeature) {
		return diag.Errorf("feature name cannot be updated. Create a new resource instead for a new IPv6 allowlist feature")
	}

	if len(deleteSubnets) > 0 {
		if err := WaitIPv6AllowlistDelete(ctx, acsClient, stack, v2.Feature(feature), deleteSubnets); err != nil {
			// if feature not found set id of resource to empty string to remove from state
			if errors.IsUnknownFeatureError(err) {
				tflog.Info(ctx, fmt.Sprintf("Invalid IPv6 Allowlist feature (%s): %s.", feature, err))
				return nil //if we return an error here, the set id will not take effect and state will be preserved
			}
			return diag.Errorf("Error updating ipv6 allowlist (%s): %s", feature, err)
		}
	}

	if len(addSubnets) > 0 {
		if err := WaitIPv6AllowlistCreate(ctx, acsClient, stack, v2.Feature(feature), addSubnets); err != nil {
			return diag.Errorf("Error updating ipv6 allowlist (%s): %s", feature, err)
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Updated IPv6 Allowlist resource for feature: %s\n", feature))

	return resourceIPv6AllowlistRead(ctx, d, m)
}

func resourceIPv6AllowlistDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Determine the changes to the subnets for a feature
	feature, oldSubnetsSet, _ := parseIPv6AllowlistRequest(d)
	deleteSubnets := utils.GetSubnetsFromSet(oldSubnetsSet)

	if len(deleteSubnets) > 0 {
		if err := WaitIPv6AllowlistDelete(ctx, acsClient, stack, v2.Feature(feature), deleteSubnets); err != nil {
			// if feature not found set id of resource to empty string to remove from state
			if errors.IsUnknownFeatureError(err) {
				tflog.Info(ctx, fmt.Sprintf("Invalid IPv6 Allowlist feature (%s): %s.", feature, err))
				return nil //if we return an error here, the set id will not take effect and state will be preserved
			}
			return diag.Errorf("Error deleting ipv6 allowlist (%s): %s", feature, err)
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Deleted IPv6 Allowlist resource for feature: %s\n", feature))
	return nil
}

func parseIPv6AllowlistRequest(d *schema.ResourceData) (feature string, oldSubnets *schema.Set, newSubnets *schema.Set) {
	feature = d.Get(schemaKeyFeature).(string)

	rawOriginalSubnets, rawNewSubnets := d.GetChange(schemaKeySubnets)
	oldSubnets = rawOriginalSubnets.(*schema.Set)
	newSubnets = rawNewSubnets.(*schema.Set)
	return feature, oldSubnets, newSubnets
}
