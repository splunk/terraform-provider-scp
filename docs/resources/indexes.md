--- 

# scp_indexes (Resource)

Index Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageIndexes for more detailed information on attribute requirements and the ACS Indexes API.

## Example Usage

```terraform
resource "scp_indexes" "index-1" {
  name = "index-1"
}
 
resource "scp_indexes" "index-2" {
  name = "index-2"
  searchable_days = 90
  lifecycle {
    prevent_destroy = true
  }
}
 
resource "scp_indexes" "index-3" {
  name = "index-3"
  searchable_days = 90
  max_data_size_mb = 512
  lifecycle {
    prevent_destroy = true
  }
}
```

## Schema

### Required

- `name` (String) The name of the index to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old index and recreate with new name). Use the lifecycle `prevent_destroy` meta-argument to prevent deletion if this field is changed.

### Optional

-  `datatype` (String) Valid values: (event | metric). Specifies the type of index. Defaults to event. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete current index and recreate with new datatype). Use the lifecycle `prevent_destroy` meta-argument to prevent deletion if this field is changed. 
-  `max_data_size_mb` (Number) The maximum size of the index in megabytes. Defaults to 0 (unlimited).
-  `searchable_days` (Number) Number of days after which indexed data rolls to frozen. Defaults to 90 days.
-  `self_storage_bucket_path` (String) To create an index with DDSS enabled, you must specify the selfStorageBucketPath value in the following format: `s3://selfStorageBucket/selfStorageBucketFolder`, where SelfStorageBucketFolder is optional, as you can store data buckets at root. Before you can create an index with DDSS enabled, you must configure a self-storage location for your deployment (see https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageDDSSlocations). Can not be set with splunk_archival_retention_days. 
-  `splunk_archival_retention_days` (Number) To create an index with DDAA enabled, you must specify the `splunk_archival_retention_days` value which must be positive and greater than the `searchable_days` value. Can not be set with `self_storage_bucket_path`.

### Read-Only

- `id` (String) The ID of this resource.

### Note 
- Changing `name` and/or `datatype` will cause the index to be destroyed and recreated. If you would like to ensure that an index is not deleted as a result of either of these fields being updated, like the lifecycle meta-argument as follows: 
```terraform
resource "scp_indexes" "index-1" {
  name = "index-1"
  lifecycle {
    prevent_destroy = true
  }
}
``` 
- Can only set either `self_storage_bucket_path` or `splunk_archival_retention_days`

## Timeouts 
Defaults are currently set to:
- `create` -  20m
- `read` -  20m
- `update` -  20m
- `delete` -  20m

## Notes/Troubleshooting

### Terraform Import 
**Issue:** If you receive a 409 conflict error when creating a resource, either use a different index name to create a new resource, or use `terraform import` to bring
  the resource under terraform management. 

**Solution:** User must manually write resource block in the config file:
```
resource "scp_indexes" "index-1" {
    #configuration here
}
```

User must then run the following command to bring the index into terraform state: 

```terraform import scp_indexes.index-1 index-1```

Note: Terraform import does NOT write the resource configuration in the `resource.tf` configuration file, only brings it
into terraform state (`.tfstate`). `.tfstate ` contains Terraform's understanding of the infrastructure/resources that are managed. 

### Remove from State 
If an index is deleted outside terraform, the provider should gracefully handle this and recreate it as long as it is still in the configuration file. 
If you wish to remove an index from terraform state entirely, you may use the following command: 

``` terraform state rm scp_indexes.index-1 ```

### Resource Replacement 
**NOTE:** If you do not want an index to be deleted in the event that an edit is made to index `name` or `datatype` fields, please include the `prevent_destroy`
field in the .tf resource block as shown in the examples at the top. 

**Issue**: Due to the async nature of ACS APIs, if replacing resources (creating and deleting resource with same name), you may 
encounter timeout on the delete operation as the poll to verify index has been deleted fails due to the newly created resource 
with the same name. 

**Solution**: Rerun `terraform apply` or run `terraform apply` with `-parallelism=1` to avoid this issue by limiting the number of simultaneous resource operations to 1 (instead of the default of 10 resources at a time)
                          
