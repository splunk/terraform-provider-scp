---


# splunkcloud Provider


## Example Usage

```terraform
provider "splunkcloud" {
  stack = "example-stack"
  server = "https://admin.splunk.com"
}
```

## Schema

### Optional

- `server` (String) ACS API base URL. May also be provided via ACS_SERVER environment variable.
- `stack` (String) Stack to perform ACS operations. May also be provided via SPLUNK_STACK environment variable.
- `auth_token` (String, Sensitive) Authentication tokens, also known as JSON Web Tokens (JWT), are a method for authenticating Splunk platform users into the Splunk platform. May also be provided via STACK_TOKEN environment variable.
- `username` (String) Splunk Cloud Platform deployment username. May also be provided via STACK_USERNAME environment variable.
- `password` (String, Sensitive) Splunk Cloud Platform deployment password. May also be provided via STACK_PASSWORD environment variable.

### NOTE:
Please note that “optional” here refers to optional to be set in the provider configuration file, not that it is truly optional to provide at all to Terraform. 

It is required that the following attributes instead be set through environment variables as specified above. Sensitive attributes such as `auth_token` and `password` should NOT be set through the configuration file in plain text. 
Instead, it is recommended to using a dedicated secret store such as Vault or AWS Secrets Manager. 
- `server`
- `stack`
- Either `auth_token` or `username`/`password` 