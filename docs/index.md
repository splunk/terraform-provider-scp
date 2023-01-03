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

- `server` (String) ACS API endpoint. May also be provided via ACS_SERVER environment variable.
- `stack` (String) Stack to perform ACS operations. May also be provided via SPLUNK_STACK environment variable.
- `auth_token` (String, Sensitive) Authentication tokens, also known as JSON Web Tokens (JWT), are a method for authenticating Splunk platform users into the Splunk platform. May also be provided via SPLUNK_AUTH_TOKEN environment variable.
- `username` (String) Splunk Cloud Platform deployment username. May also be provided via STACK_USERNAME environment variable.
- `password` (String, Sensitive) Splunk Cloud Platform deployment password. May also be provided via STACK_PASSWORD environment variable.

### NOTE
**Please note that “optional” here refers to optional to be set in the provider configuration file, not that it is truly optional to provide at all to Terraform. 

It is required that the following attributes instead be set through environment variables as specified above, and sensitive attributes such as `auth_token` and `password` should be set through an environment variable.
- `server`
- `stack`
- Either `auth_token` or `username`/`password` 