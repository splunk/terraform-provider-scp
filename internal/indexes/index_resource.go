package indexes

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"net/http"
	"strings"
)

func ResourceIndex() *schema.Resource {
	return &schema.Resource{
		// This description is used by the documentation generator and the language server.
		Description: "Index Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/" +
			"ManageIndexes for more latest, detailed information on attribute requirements and the ACS Indexes API. ",

		CreateContext: resourceIndexCreate,
		ReadContext:   resourceIndexRead,
		UpdateContext: resourceIndexUpdate,
		DeleteContext: resourceIndexDelete,
		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: IndexSchema(),
	}
}

func IndexSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"name": {
			Type:        schema.TypeString,
			Required:    true,
			ForceNew:    true,
			Description: "The name of the index to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old index and recreate with new name).",
		},
		"datatype": {
			Type:        schema.TypeString,
			Optional:    true,
			Computed:    true,
			ForceNew:    true,
			Description: "Valid values: (event | metric). Specifies the type of index. Can not be updated. Defaults to event. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete current index and recreate with new datatype).\n",
		},
		"max_data_size_mb": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Computed:    true,
			Description: "The maximum size in MB for a hot DB to reach before a roll to warm is triggered. Defaults to 0 (unlimited)",
		},
		"searchable_days": {
			Type:        schema.TypeFloat,
			Optional:    true,
			Computed:    true,
			Description: "Number of days after which indexed data rolls to frozen. Defaults to 90 days",
		},
		"self_storage_bucket_path": {
			Type:          schema.TypeString,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"splunk_archival_retention_days"},
			Description:   "To create an index with DDSS enabled, you must specify the selfStorageBucketPath value in the following format: \"s3://selfStorageBucket/selfStorageBucketFolder\", where SelfStorageBucketFolder is optional, as you can store data buckets at root. Before you can create an index with DDSS enabled, you must configure a self-storage location for your deployment. Can not be set with splunk_archival_retention_days ",
		},
		"splunk_archival_retention_days": {
			Type:          schema.TypeFloat,
			Optional:      true,
			Computed:      true,
			ConflictsWith: []string{"self_storage_bucket_path"},
			Description:   "To create an index with DDAA enabled, you must specify the splunkArchivalRetentionDays value which must be The value of splunkArchivalRetentionDays must be positive and greater than or equal to the SearchableDays value. Can not be set with self_storage_bucket_path",
		},
	}
}

func resourceIndexCreate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client and stack from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	// Retrieve data for each field and create request body
	indexRequest := parseIndexRequest(d)
	createIndexRequest := v2.CreateIndexJSONRequestBody{
		Datatype:                    indexRequest.Datatype,
		Name:                        indexRequest.Name,
		MaxDataSizeMB:               indexRequest.MaxDataSizeMB,
		SearchableDays:              indexRequest.SearchableDays,
		SplunkArchivalRetentionDays: indexRequest.SplunkArchivalRetentionDays,
		SelfStorageBucketPath:       indexRequest.SelfStorageBucketPath,
	}

	tflog.Info(ctx, fmt.Sprintf("%+v\n", createIndexRequest))

	err := WaitIndexCreate(ctx, acsClient, stack, createIndexRequest)
	if err != nil {
		if stateErr := err.(*resource.UnexpectedStateError); stateErr.State == http.StatusText(http.StatusConflict) {
			return diag.Errorf(fmt.Sprintf("Index (%s) already exists, use a different name to create index or use terraform import to bring current index under terraform management", indexRequest.Name))
		}

		return diag.Errorf(fmt.Sprintf("Error submitting request for index (%s) to be created: %s", indexRequest.Name, err))
	}

	// Poll Index until GET returns 200 to confirm index creation
	err = WaitIndexPoll(ctx, acsClient, stack, indexRequest.Name, TargetStatusResourceExists, PendingStatusVerifyCreated)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for index (%s) to be created: %s", indexRequest.Name, err))
	}

	// Set ID of index resource to indicate index has been created
	d.SetId(indexRequest.Name)

	tflog.Info(ctx, fmt.Sprintf("Created index resource: %s\n", indexRequest.Name))

	// Call readIndex to set attributes of index
	return resourceIndexRead(ctx, d, m)
}

func resourceIndexRead(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve your client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	indexName := d.Id()

	index, err := WaitIndexRead(ctx, acsClient, stack, indexName)

	if err != nil {
		// if index not found set id of resource to empty string to remove from state
		if stateErr := err.(*resource.UnexpectedStateError); strings.Contains(stateErr.LastError.Error(), "404-index-not-found") {
			tflog.Info(ctx, fmt.Sprintf("Removing index from state. Not Found error while reading index (%s): %s.", indexName, err))
			d.SetId("")
			return nil //if we return an error here, the set id will not take effect and state will be preserved
		} else {
			return diag.Errorf(fmt.Sprintf("Error reading index (%s): %s", indexName, err))
		}
	}

	if err := d.Set("name", d.Id()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("datatype", index.Datatype); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("max_data_size_mb", index.MaxDataSizeMB); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("searchable_days", index.SearchableDays); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("self_storage_bucket_path", index.SelfStorageBucketPath); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("splunk_archival_retention_days", index.SplunkArchivalRetentionDays); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func resourceIndexUpdate(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	indexName := d.Id()

	// Retrieve data for each field and create request body
	indexRequest := parseIndexRequest(d)
	patchRequest := v2.PatchIndexInfoJSONRequestBody{
		MaxDataSizeMB:               indexRequest.MaxDataSizeMB,
		SearchableDays:              indexRequest.SearchableDays,
		SplunkArchivalRetentionDays: indexRequest.SplunkArchivalRetentionDays,
		SelfStorageBucketPath:       indexRequest.SelfStorageBucketPath,
	}

	err := WaitIndexUpdate(ctx, acsClient, stack, patchRequest, indexName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error submitting request for index (%s) to be updated: %s", indexName, err))
	}

	//Poll until fields have been confirmed updated
	err = WaitIndexConfirmUpdate(ctx, acsClient, stack, patchRequest, indexName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for index (%s) to be updated: %s", indexName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("updated index resource: %s\n", indexName))
	return resourceIndexRead(ctx, d, m)
}

func resourceIndexDelete(ctx context.Context, d *schema.ResourceData, m interface{}) diag.Diagnostics {
	// use the meta value to retrieve client from the provider configure method
	acsProvider := m.(client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	indexName := d.Id()

	err := WaitIndexDelete(ctx, acsClient, stack, indexName)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error deleting index (%s): %s", indexName, err))
	}

	//Poll Index until GET returns 404 Not found - index has been deleted
	err = WaitIndexPoll(ctx, acsClient, stack, indexName, TargetStatusResourceDeleted, PendingStatusVerifyDeleted)
	if err != nil {
		return diag.Errorf(fmt.Sprintf("Error waiting for index (%s) to be deleted: %s", indexName, err))
	}

	tflog.Info(ctx, fmt.Sprintf("deleted index resource: %s\n", indexName))
	return nil
}

func parseIndexRequest(d *schema.ResourceData) *v2.IndexInfo {
	indexRequest := v2.IndexInfo{}

	indexRequest.Name = d.Get("name").(string)

	if dataType, ok := d.GetOk("datatype"); ok {
		parsedDataType := dataType.(string)
		indexRequest.Datatype = &parsedDataType
	}

	if maxDataSizeMB, ok := d.GetOk("max_data_size_mb"); ok {
		parsedData := int64(maxDataSizeMB.(float64))
		indexRequest.MaxDataSizeMB = &parsedData
	}

	if searchableDays, ok := d.GetOk("searchable_days"); ok {
		parsedData := int64(searchableDays.(float64))
		indexRequest.SearchableDays = &parsedData
	}

	if bucketPath, ok := d.GetOk("self_storage_bucket_path"); ok {
		parsedData := bucketPath.(string)
		indexRequest.SelfStorageBucketPath = &parsedData
	}

	if retentionDays, ok := d.GetOk("splunk_archival_retention_days"); ok {
		parsedData := int64(retentionDays.(float64))
		indexRequest.SplunkArchivalRetentionDays = &parsedData
	}

	return &indexRequest
}
