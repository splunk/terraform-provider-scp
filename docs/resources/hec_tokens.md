--- 

# scp_hec_tokens (Resource)

Hec Token Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageHecTokens for more detailed information on attribute requirements and the ACS HEC Token API.

## Example Usage

```terraform
resource "scp_hec_tokens" "hec-1" {
  name = "hec-1"
}
 
resource "scp_hec_tokens" "hec-2" {
  name = "hec-2"
  allowed_indexes = ["main", "summary"]
  default_index = "main" 
  use_ack = true 
  disabled = false
  lifecycle {
    prevent_destroy = true
  }
}

```

## Schema

### Required

- `name` (String) The name of the hec token to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old hec token and recreate with new name)

### Optional

- `allowed_indexes` (Set of String) Set of indexes allowed for events with this token
- `default_index` (String) Index to store generated events. Must not be an empty string. If allowed_indexes is provided, default_index must be one of allowed_indexes  
- `default_source` (String) Default source for events with this token
- `default_sourcetype` (String) Default sourcetype for events with this token
- `disabled` (Boolean) Input disabled indicator: false = Input Not disabled, true = Input disabled
- `token` (String) Token value for sending data to collector/event endpoint. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old HEC and recreate with new token). Use the lifecycle `prevent_destroy` meta-argument to prevent deletion if this field is changed.
- `use_ack` (Boolean) Indexer acknowledgement for this token: false = disabled, true = enabled

### Read-Only

- `id` (String) The ID of this resource.

### Note
- Changing `name` and/or `token` will cause the hec token to be destroyed and recreated. If you would like to ensure that a hec token is not deleted as a result of either of these fields being updated, like the lifecycle meta-argument as follows:
```terraform
resource "scp_hec_tokens" "hec-1" {
  name = "hec-1"
  lifecycle {
    prevent_destroy = true
  }
}
``` 

## Timeouts
Defaults are currently set to:
- `create` -  20m
- `read` -  20m
- `update` -  20m
- `delete` -  20m 

## Notes/Troubleshooting

### Terraform Import
**Issue:** If you receive a 409 conflict error when creating a resource, either use a different hec name to create a new resource, or use `terraform import` to bring
the resource under terraform management.

**Solution:** User must manually write resource block in the config file:
```
resource "scp_hec_tokens" "hec-1" {
    #configuration here
}
```

User must then run the following command to bring the hec into terraform state:

```terraform import scp_hec_tokens.hec-1 hec-1```

Note: Terraform import does NOT write the resource configuration in the `resource.tf` configuration file, only brings it
into terraform state (`.tfstate`). `.tfstate ` contains Terraform's understanding of the infrastructure/resources that are managed.

### Remove from State
If a hec token is deleted outside terraform, the provider should gracefully handle this and recreate it as long as it is still in the configuration file.
If you wish to remove a hec from terraform state entirely, you may use the following command:

``` terraform state rm scp_hec_tokens.hec-1 ```

### Resource Replacement
**NOTE:** If you do not want a hec token to be deleted in the event that an edit is made to hec `name` or `datatype` fields, please include the `prevent_destroy`
field in the .tf resource block as shown in the examples at the top.

**Issue**: Due to the async nature of ACS APIs, if replacing resources (creating and deleting resource with same name), you may
encounter timeout on the delete operation as the poll to verify hec has been deleted fails due to the newly created resource
with the same name.

**Solution**: Rerun `terraform apply` or run `terraform apply` with `-parallelism=1` to avoid this issue by limiting the number of simultaneous resource operations to 1 (instead of the default of 10 resources at a time)
