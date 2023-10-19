****---


# Splunk Cloud Platform (scp) Provider


## Example Usage

```terraform
provider "scp" {
  stack = "example-stack"
  server = "https://admin.splunk.com"
}
```

### NOTE:
Please note that it is optional to set the provider attributes in the provider configuration file, and may also be passed as an environmental variable.

Sensitive attributes such as `auth_token` and `password` should NOT be set through the configuration file in plain text.
Instead, it is recommended to using a dedicated secret store such as Vault or AWS Secrets Manager.

### Required Attributes
The following attributes must be set for the provider to work.
- `server`
- `stack`
- Either `auth_token` or `username`/`password` NOTE: IL2 environment will not be able to use `username`/`password` for authentication.

## Schema

- `server` (String) ACS API base URL. May also be provided via ACS_SERVER environment variable.
- `stack` (String) Stack to perform ACS operations. May also be provided via SPLUNK_STACK environment variable.
- `auth_token` (String, Sensitive) Authentication tokens, also known as JSON Web Tokens (JWT), are a method for authenticating Splunk platform users into the Splunk platform. May also be provided via STACK_TOKEN environment variable.
- `username` (String) Splunk Cloud Platform deployment username. May also be provided via STACK_USERNAME environment variable.
- `password` (String, Sensitive) Splunk Cloud Platform deployment password. May also be provided via STACK_PASSWORD environment variable.

## Configuring Stack Deployment: Special Cases 

### Targeting A Search Head

This provider supports passing in a search head prefix to target the CRUD operations ran by Terraform on a specific search head. 
Please refer to [ACS API Documentation](https://docs.splunk.com/Documentation/SplunkCloud/9.0.2205/Config/ACSIntro) to understand 
targeting implications and limitations by feature. Not all features support targeting a specific search head, 
so it is not advised to rely solely on a search head in your provider configuration. 

Please note that to target a search head and use token authentication, `auth_token` must be set to a token that was created
on that specific search head. 

#### Example

```terraform
provider "scp" {
  stack = "sh-i-0112a21f78ba1c3.example-stack"
  server = "https://admin.splunk.com"
  auth_token = var.sh_token
}
```

### Managing Multiple Deployments 

Through this Terraform provider, users can manage multiple types of deployments, such as Victoria/Classic stacks and targeting search heads through the 
provider configuration resource block. This allows managing multiple stacks, multiple types of stacks, and also 
managing certain resources on specific search heads. Please refer to the [ACS API Compatibility matrix](https://docs.splunk.com/Documentation/SplunkCloud/9.0.2205/Config/ACSreqs) to understand which
features are supported for your Stack Deployment experience. 

#### Example 

```terraform
variable "victoria_token" {
  description = "The auth token for Victoria deployment"
  type        = string
}

variable "victoria_sh1_token" {
  description = "The auth token for Victoria targeted sh"
  type        = string
}

variable "classic_token" {
  description = "The auth token for Classic deployment"
  type        = string
}

provider "scp" {
  stack = "primary-victoria-stack"
  server = "https://admin.splunk.com"
  auth_token = var.victoria_token
}

provider "scp" {
  alias = "victoria-sh1"
  stack = "sh-i-0p1b8c294321d1b8.primary-victoria-stack"
  server = "https://admin.splunk.com"
  auth_token = var.victoria_sh1_token
}

provider "scp" {
  alias = "classic"
  stack = "classic-stack"
  server = "https://admin.splunk.com"
  auth_token = var.classic_token
}

resource "scp_indexes" "victoria-index-1" {
  provider = scp
  name = "victoria-index-1"
}

resource "scp_indexes" "sh-index-1" {
  provider = scp.victoria-sh1
  name = "sh-index-1"
  searchable_days = 50
  splunk_archival_retention_days = 100
}

resource "scp_indexes" "classic-index-1" {
  provider = scp.classic
  name = "classic-index-1"
}
```
## General Notes/Troubleshooting

### Classic vs Victoria

While the Terraform provider supports both Classic and Victoria deployments, there are some limitations on Classic. 
First, user should run `terraform apply` with `-parallelism=1` flag to prevent concurrent write operations. This results in 
Classic `terraform apply` runs taking longer than Victoria deployments. Second, when managing Hec Token resource, 
search head targeting may not be used (see above for examples on how to target a search head for some resources and not others.)

### Retries

The Terraform provider is configured to retry on certain error codes from the ACS API, such as error code 429 caused by
ACS API rate limiting. When hitting a rate limit, it will likely take about 5 minutes for requests to become accepted again.

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


