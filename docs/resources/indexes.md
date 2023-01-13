--- 

# splunkcloud_indexes (Resource)

Index Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageIndexes for more latest, detailed information on attribute requirements and the ACS Indexes API.

## Example Usage

```terraform
resource "splunkcloud_indexes" "index-1" {
  name = "index-1"
}
 
resource "splunkcloud_indexes" "index-2" {
  name = "index-2"
  searchable_days = 90
}
 
resource "splunkcloud_indexes" "index-3" {
  name = "index-3"
  searchable_days = 90
  max_data_size_mb = 512
}
```

## Schema

### Required

- `name` (String) The name of the index to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old index and recreate with new name).

### Optional

-  `datatype` (String) Valid values: (event | metric). Specifies the type of index. Defaults to event. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete current index and recreate with new datatype).
-  `max_data_size_mb` (Number) The maximum size in MB for a hot DB to reach before a roll to warm is triggered. Defaults to 0 (unlimited)
-  `searchable_days` (Number) Number of days after which indexed data rolls to frozen. Defaults to 90 days
-  `self_storage_bucket_path` (String) To create an index with DDSS enabled, you must specify the selfStorageBucketPath value in the following format: "s3://selfStorageBucket/selfStorageBucketFolder", where SelfStorageBucketFolder is optional, as you can store data buckets at root. Before you can create an index with DDSS enabled, you must configure a self-storage location for your deployment. Can not be set with splunk_archival_retention_days
-  `splunk_archival_retention_days` (Number) To create an index with DDAA enabled, you must specify the splunkArchivalRetentionDays value which must be The value of splunkArchivalRetentionDays must be positive and greater than or equal to the SearchableDays value. Can not be set with self_storage_bucket_path

### Read-Only

- `id` (String) The ID of this resource.

### Note 
- Changing `name` and/or `datatype` will cause the index to be destroyed and recreated.
- Can only set either `self_storage_bucket_path` or `splunk_archival_retention_days`

## Timeouts 
Defaults are currently set to:
- `create` -  20m
- `update` -  20m
- `update` -  20m
- `delete` -  20m 

## Notes/Troubleshooting 

### Retries 

The Terraform provider is configured to retry on certain error codes from the ACS API, such as error code 429, due 
to the ACS API rate limiting. When hitting a rate limit, it will likely take about 5 minutes for requests to become accepted again. 

### Terraform Import 
**Issue:** If you receive a 409 conflict error when creating a resource, either use a different index name to create a new resource, or use `terraform import` to bring
  the resource under terraform management. 

**Solution:** User must manually write resource block in the config file:
```
resource "splunkcloud_indexes" "index-1" {
    #configuration here
}
```

User must then run the following command to bring the index into terraform state: 

```terraform import splunkcloud_indexes.index-1 index-1```

Note: Terraform import does NOT write the resource configuration in the config file, only brings it into terraform state
                
### Remove from State 
If an index is deleted outside terraform, the provider should gracefully handle this and recreate it as long as it is still in the configuration file. 
If you wish to remove an index from terraform state entirely, you may use the following command: 

``` terraform state rm splunkcloud_indexes.index-1 ```

### Resource Replacement 
**Issue**: Due to the async nature of ACS APIs, if replacing resources (creating and deleting resource with same name), you may 
encounter timeout on the delete operation as the poll to verify index has been deleted fails due to the newly created resource 
with the same name. 

**Solution**: Rerun `terraform apply` or run with `-parallelism=1` to avoid this issue. 
                          
### Errors from the ACS API: 
Unexpected errors received from the ACS API such as bad requests will be output to the user as indicated below. 

Please see https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ACSerrormessages for general troubleshooting tips: 

``` 
Error submitting request for index (index-1) to be created: 
unexpected state 'Not Found', wanted target 'Accepted'. 
last error: {"code":"404-stack-not-found","message":"stack not found. 
Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ACSerrormessages 
for general troubleshooting tips."}
```

### Logs 
Please see the following for more information on viewing logs for terraform provider. https://developer.hashicorp.com/terraform/plugin/log/managing
