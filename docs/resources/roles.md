--- 

# scp_roles (Resource)

Role Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageRoles for more detailed information on attribute requirements and the ACS Roles API.

## Example Usage

```terraform
resource "scp_roles" "log-viewer-role" {
  name = "log-viewer-role"
}

resource "scp_roles" "app-manager-role" {
  name = "app-manager-role"
  capabilities                  = ["accelerate_search"]
  cumulative_rt_srch_jobs_quota = 200
  cumulative_srch_jobs_quota    = 100
  default_app                   = "search"
  imported_roles                = ["user"]
  rt_srch_jobs_quota            = 20
  srch_jobs_quota               = 10
  srch_disk_quota               = 500
  srch_filter                   = "*"
  srch_time_earliest            = -1
  srch_time_win                 = -1
  federated_search_manage_ack   = "Y"
}
```

## Schema

### Required

- `name` (String) The name of the role to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old role and recreate with new name). Use the lifecycle `prevent_destroy` meta-argument to prevent deletion if this field is changed.

### Optional

-  `capabilities` (Set of String) The capabilities attached to the role.
-  `cumulative_rt_srch_jobs_quota` (Number) Maximum number of concurrently running real-time searches that all members of this role can have. The value must be a non-negative number.
-  `cumulative_srch_jobs_quota` (Number) Maximum number of concurrently running historical searches that all members of this role can have. The value must be a non-negative number.
-  `default_app` (String) The default app for this role.
-  `imported_roles` (Set of String) List of other roles and their associated capabilities that should be imported.
-  `rt_srch_jobs_quota` (Number) Maximum number of concurrently running real-time searches a member of this role can have. The value must be a non-negative number.
-  `srch_jobs_quota` (Number) Maximum number of concurrently running historical searches a member of this role can have. The value must be a non-negative number.
-  `srch_disk_quota` (Number) Maximum amount of disk space (MB) that can be used by search jobs of a user that belongs to this role. The value must be a non-negative number.
-  `srch_filter` (String) Search filters for this Role.
-  `srch_indexes_allowed` (Set of String) List of indexes this role is allowed to search.
-  `srch_indexes_default` (Set of String) List of indexes to search when no index is specified.
-  `srch_time_earliest` (Number) Maximum amount of time that searches of users from this role will be allowed to run. A value of -1 means unset, 0 means infinite. Any other value is the amount of time in seconds, for example, 300 would mean 300s.
-  `srch_time_win` (Number) Maximum time span of a search, in seconds. A value of -1 means unset, 0 means infinite. Any other value is the amount of time in seconds, for example, 300 would mean 300s.
-  `federated_search_manage_ack` (String) If 'imported_roles' or 'capabilities' contains the 'fsh_manage' capability, you must set this attribute to a value of "Y". This header acknowledges that a role with the 'fsh_manage' capability can send search results outside the compliant environment.

### Read-Only

- `id` (String) The ID of this resource.

### Note 
- Changing `name` will cause the role to be destroyed and recreated. If you would like to ensure that a role is not deleted as a result of either of these fields being updated, like the lifecycle meta-argument as follows: 
```terraform
resource "scp_roles" "role-1" {
  name = "role-1"
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
**Issue:** If you receive a 409 conflict error when creating a resource, either use a different role name to create a new resource, or use `terraform import` to bring
  the resource under terraform management. 

**Solution:** User must manually write resource block in the config file:
```
resource "scp_roles" "role-1" {
    #configuration here
}
```

User must then run the following command to bring the role into terraform state: 

```terraform import scp_roles.role-1 role-1```

Note: Terraform import does NOT write the resource configuration in the `resource.tf` configuration file, only brings it
into terraform state (`.tfstate`). `.tfstate ` contains Terraform's understanding of the infrastructure/resources that are managed. 

### Remove from State 
If a role is deleted outside terraform, the provider should gracefully handle this and recreate it as long as it is still in the configuration file. 
If you wish to remove a role from terraform state entirely, you may use the following command: 

``` terraform state rm scp_roles.role-1 ```

### Resource Replacement 
**NOTE:** If you do not want a role to be deleted in the event that an edit is made to the role `name` field, please include the `prevent_destroy`
field in the .tf resource block as shown in the examples at the top.

### Search Head Targeting
Some ACS operations including Users and Roles are not replicated across search heads and require search head targeting. Here is an example of specifying search head targeting in the provider. https://github.com/splunk/terraform-provider-scp/blob/main/docs/index.md#targeting-a-search-head
Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/TargetSearchHeads for more detailed information on how to configure search head targeting in ACS.
