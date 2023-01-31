---


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
- Either `auth_token` or `username`/`password`

## Schema

- `server` (String) ACS API base URL. May also be provided via ACS_SERVER environment variable.
- `stack` (String) Stack to perform ACS operations. May also be provided via SPLUNK_STACK environment variable.
- `auth_token` (String, Sensitive) Authentication tokens, also known as JSON Web Tokens (JWT), are a method for authenticating Splunk platform users into the Splunk platform. May also be provided via STACK_TOKEN environment variable.
- `username` (String) Splunk Cloud Platform deployment username. May also be provided via STACK_USERNAME environment variable.
- `password` (String, Sensitive) Splunk Cloud Platform deployment password. May also be provided via STACK_PASSWORD environment variable.

