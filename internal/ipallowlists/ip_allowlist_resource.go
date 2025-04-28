package ipallowlists

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/errors"
)

const (
	ResourceKey = "scp_ip_allowlists"

	schemaKeyFeature = "feature"
	schemaKeySubnets = "subnets"
)

func ipAllowlistResourceSchema() map[string]*schema.Schema {
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
			Description: "Subnets is a list of IP addresses that have access to the corresponding feature.",
			MinItems:    1,
		},
	}
}

func ResourceIPAllowlist() *schema.Resource {
	return &schema.Resource{
		Description: "IP Allowlist Resource. Please see documentation to understand unique behavior regarding naming and delete operation. " +
			"Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ConfigureIPAllowList " +
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
	feature, _, newSubnetsSet := parseIPAllowlistRequest(d)
	addSubnets := GetSubnetsFromSet(newSubnetsSet)

	// Add new subnets
	err := WaitIPAllowlistCreate(ctx, acsClient, stack, v2.Feature(feature), addSubnets)
	if err != nil {
		if errors.IsUnknownFeatureError(err) {
			tflog.Info(ctx, fmt.Sprintf("Invalid IP Allowlist feature (%s): %s.", feature, err))
		}
		return diag.Errorf(fmt.Sprintf("Error submitting request for IP allowlist (%s) to be created: %s", feature, err))
	}

	// Set ID of index resource to indicate index has been created
	d.SetId(feature)
	tflog.Info(ctx, fmt.Sprintf("Created IP Allowlist resource for feature: %s\n", feature))

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
		if errors.IsUnknownFeatureError(err) {
			tflog.Info(ctx, fmt.Sprintf("Invalid IP Allowlist feature (%s): %s.", feature, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		}
		return diag.Errorf(fmt.Sprintf("Error reading ip allowlist (%s): %s", feature, err))
	}

	if err := d.Set(schemaKeyFeature, d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set(schemaKeySubnets, subnets); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceIPAllowlistUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Determine the changes to the subnets for a feature
	feature, oldSubnetsSet, newSubnetsSet := parseIPAllowlistRequest(d)
	addSubnets := GetSubnetsFromSet(newSubnetsSet.Difference(oldSubnetsSet))
	deleteSubnets := GetSubnetsFromSet(oldSubnetsSet.Difference(newSubnetsSet))

	// do not allow feature name to be changed
	if d.HasChange(schemaKeyFeature) {
		return diag.Errorf("feature name cannot be updated. Create a new resource instead for a new IP allowlist feature")
	}

	if len(deleteSubnets) > 0 {
		if err := WaitIPAllowlistDelete(ctx, acsClient, stack, v2.Feature(feature), deleteSubnets); err != nil {
			// if feature not found set id of resource to empty string to remove from state
			if errors.IsUnknownFeatureError(err) {
				tflog.Info(ctx, fmt.Sprintf("Invalid IP Allowlist feature (%s): %s.", feature, err))
				return nil //if we return an error here, the set id will not take effect and state will be preserved
			}
			return diag.Errorf(fmt.Sprintf("Error updating ip allowlist (%s): %s", feature, err))
		}
	}

	if len(addSubnets) > 0 {
		if err := WaitIPAllowlistCreate(ctx, acsClient, stack, v2.Feature(feature), addSubnets); err != nil {
			return diag.Errorf(fmt.Sprintf("Error updating ip allowlist (%s): %s", feature, err))
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Updated IP Allowlist resource for feature: %s\n", feature))

	return resourceIPAllowlistRead(ctx, d, m)
}

func resourceIPAllowlistDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Determine the changes to the subnets for a feature
	feature, oldSubnetsSet, _ := parseIPAllowlistRequest(d)
	deleteSubnets := GetSubnetsFromSet(oldSubnetsSet)

	if len(deleteSubnets) > 0 {
		if err := WaitIPAllowlistDelete(ctx, acsClient, stack, v2.Feature(feature), deleteSubnets); err != nil {
			// if feature not found set id of resource to empty string to remove from state
			if errors.IsUnknownFeatureError(err) {
				tflog.Info(ctx, fmt.Sprintf("Invalid IP Allowlist feature (%s): %s.", feature, err))
				return nil //if we return an error here, the set id will not take effect and state will be preserved
			}
			return diag.Errorf(fmt.Sprintf("Error deleting ip allowlist (%s): %s", feature, err))
		}
	}

	tflog.Info(ctx, fmt.Sprintf("Deleted IP Allowlist resource for feature: %s\n", feature))
	return nil
}

func parseIPAllowlistRequest(d *schema.ResourceData) (feature string, oldSubnets *schema.Set, newSubnets *schema.Set) {
	feature = d.Get(schemaKeyFeature).(string)

	rawOriginalSubnets, rawNewSubnets := d.GetChange(schemaKeySubnets)
	oldSubnets = rawOriginalSubnets.(*schema.Set)
	newSubnets = rawNewSubnets.(*schema.Set)
	return feature, oldSubnets, newSubnets
}

func GetSubnetsFromSet(subnets *schema.Set) []string {
	result := make([]string, 0)
	if subnets == nil {
		return result
	}
	for _, subnet := range subnets.List() {
		result = append(result, subnet.(string))
	}
	return result
}
