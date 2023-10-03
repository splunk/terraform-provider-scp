--- 

# scp_users (Resource)

User Resource. Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ManageRoles for more detailed information on attribute requirements and the ACS Users API.

## Example Usage

```terraform
resource "scp_users" "log-viewer" {
  name                        = "log-viewer"
}

resource "scp_users" "app-manager" {
  name                        = "app-manager"
  password                    = "insecurepassword"
  default_app                 = "launcher"
  roles                       = ["user"]
  federated_search_manage_ack = "Y"
  email                       = "tester@domain.com"
}
```

## Schema

### Required

- `name` (String) The name of the user to create. Can not be updated after creation, if changed in config file terraform will propose a replacement (delete old user and recreate with new name). Use the lifecycle `prevent_destroy` meta-argument to prevent deletion if this field is changed.

### Optional

-  `password` (String) The password of the user to create, or the new password to update. Please refer to the `Handle passwords` section on how to protect sensitive credentials.
-  `old_password` (String) The old password of the user to update. Please refer to the `Handle passwords` section on how to protect sensitive credentials.
-  `default_app` (String) Set the default app for this user. Setting this here overrides the default app inherited from the user's role(s).
-  `email` (String) The email of the user.
-  `force_change_pass` (Boolean) To force a change of password on the user's first login, set forceChangePass to "true".
-  `full_name` (String) The full name of the user.
-  `roles` (Set of String) Assign one of more roles to this user. The user will inherit all the settings and capabilities from those roles.
-  `federated_search_manage_ack` (String) If any role contains the 'fsh_manage' capability you must set this attribute to a value of "Y". This header acknowledges that a role with the fsh_manage capability can send search results outside the compliant environment.

### Read-Only

- `id` (String) The ID of this resource.
-  `last_successful_login` (string) Last successful login timestamp of the user.
-  `locked_out` (Boolean) Whether the user account has been locked out.
-  `default_app_source` (String) Default app source of the user.

### Note 
- Changing `name` will cause the user to be destroyed and recreated. If you would like to ensure that a user is not deleted as a result of either of these fields being updated, like the lifecycle meta-argument as follows: 
```terraform
resource "scp_users" "user-1" {
  name = "user-1"
  lifecycle {
    prevent_destroy = true
  }
}
``` 
- `old_password` is only needed when changing the password of the user that is associated with the JWT authentication token in the header.

## Timeouts 
Defaults are currently set to:
- `create` -  20m
- `read` -  20m
- `update` -  20m
- `delete` -  20m

## Notes/Troubleshooting

### Terraform Import 
**Issue:** If you receive a 409 conflict error when creating a resource, either use a different user name to create a new resource, or use `terraform import` to bring
  the resource under terraform management. 

**Solution:** User must manually write resource block in the config file:
```
resource "scp_users" "user-1" {
    #configuration here
}
```

User must then run the following command to bring the user into terraform state: 

```terraform import scp_users.user-1 user-1```

Note: Terraform import does NOT write the resource configuration in the `resource.tf` configuration file, only brings it
into terraform state (`.tfstate`). `.tfstate ` contains Terraform's understanding of the infrastructure/resources that are managed. 

### Remove from State 
If a user is deleted outside terraform, the provider should gracefully handle this and recreate it as long as it is still in the configuration file. 
If you wish to remove a user from terraform state entirely, you may use the following command: 

``` terraform state rm scp_users.user-1 ```

### Resource Replacement 
**NOTE:** If you do not want a user to be deleted in the event that an edit is made to the user `name` field, please include the `prevent_destroy`
field in the .tf resource block as shown in the examples at the top. 

### Manage role
The terraform user resource does not support the `createRole` attribute. In order to associate a role with the user, you must specify at least one existing role using the `roles` parameter.

### Handle password
**NOTE:** If the password is updated outside terraform (e.g. through Splunk Web), then it will cause a state divergence, because terraform will not be able to know from the remote stack that the password has been updated.

To protect your password, you can replace the credentials with variables configured with the sensitive flag, and set values for these variables using environment variables or with a .tfvars file. 
- First, declare an input variable for the password in variables.tf.
```terraform
variable "user-1-password" {
  description = "the password of user-1"
  type        = string
  sensitive   = true
}
``` 
- Second, set the password value in the .tfvars file
```terraform
  user-1-password = "insecurepassword"
```
- Last, update your user resource to reference this variable
```terraform
resource "scp_users" "user-1" {
  name                        = "user-1"
  password                    = var.user-1-password
  # more configurations
}
```
Please refer to https://developer.hashicorp.com/terraform/tutorials/configuration-language/sensitive-variables for more details.

### Search Head Targeting
Some ACS operations including Users and Roles are not replicated across search heads and require search head targeting. Here is an example of specifying search head targeting in the provider. https://github.com/splunk/terraform-provider-scp/blob/main/docs/index.md#targeting-a-search-head
Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/TargetSearchHeads for more detailed information on how to configure search head targeting in ACS.
