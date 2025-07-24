package indexes

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/splunk/terraform-provider-scp/client"
)

func indexDataSourceSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			Description: "The name of the index.",
		},
	}
}

func DataSourceIndex() *schema.Resource {
	return &schema.Resource{
		Description: "Index Data Source. Use this data source to reference default indexes " +
			"(https://docs.splunk.com/Documentation/Splunk/latest/Indexer/Aboutmanagingindexes) or other indexes " +
			"you do not wish Terraform to execute write operations on.",

		ReadContext: dataSourceIndexRead,
		Timeouts: &schema.ResourceTimeout{
			Read: schema.DefaultTimeout(20 * time.Minute),
		},
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: indexDataSourceSchema(),
	}
}

func dataSourceIndexRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Name for an index must be unique. Therefore, we read it based on the name value.
	indexName := d.Get("name").(string)

	index, err := WaitIndexRead(ctx, acsClient, stack, indexName)

	if err != nil {
		// if index not found set id of resource to empty string to remove from state
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), "404-index-not-found") {
			tflog.Info(ctx, fmt.Sprintf("Removing index from state. Not Found error while reading index (%s): %s.", indexName, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		}
		return diag.Errorf("Error reading index (%s): %s", indexName, err)
	}

	if err := d.Set("name", index.Name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(index.Name)

	return nil
}
