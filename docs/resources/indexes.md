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

