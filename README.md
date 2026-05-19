# Terraform Provider for Splunk Cloud Platform 

At this point in time, this provider supports the following resources for Splunk Cloud Platform deployments.
- Indexes
- Hec Tokens
- IP Allowlist 
- Users
- Roles
- IPv6 Allowlist
- Private App*
- Splunkbase App*

Resources marked with * are beta resources for now.

```
Copyright 2023 Splunk Inc. 

This Source Code Form is subject to the terms of the Mozilla Public License, v. 2.0. 
If a copy of the MPL was not distributed with this file, You can obtain one at 
https://www.mozilla.org/en-US/MPL/2.0/
```

## Requirements

-	[Terraform](https://www.terraform.io/downloads.html) >= 0.13.x
-	[Go](https://golang.org/doc/install) >= 1.18

## Building The Provider  

1. Clone the repository
1. Create go ```src``` directory and setup ```$GOPATH ```
1. Enter the provider directory 
1. Compile the provider by running ```make build```

## Using the Provider (Local build)

- Install Terraform
- Tell Terraform where to locate the provider 
  - With `dev_overrides`, we tell Terraform where to locate the provider locally as we will not be pulling from the registry. Make sure the path is where the provider has been compiled 
  - First `vim ~/.terraformrc` and paste the following in it:
    - ```
      provider_installation {
           dev_overrides {
               "registry.terraform.io/splunk/scp" = "<path to local terraform binary>"
           }
      }

- To update run ```terraform plan``` first to check config diff
- Run ```terraform apply``` to apply configurations
- NOTE: running `terraform init` with `dev_overrides` is not necessary and may result in unexpected errors. 

## Provider Authentication

### Stack Authentication (required)
All resources require authentication to the Splunk Cloud Platform via ACS. Provide one of:
- `auth_token` (or `STACK_TOKEN` env var) — a JWT for your stack
- `username` / `password` (or `STACK_USERNAME` / `STACK_PASSWORD` env vars) — stack deployment credentials (an ephemeral token will be generated)

### Splunkbase / AppInspect Authentication (optional)
The `splunk_username` and `splunk_password` parameters (or `SPLUNK_USERNAME` / `SPLUNK_PASSWORD` env vars) are your **splunk.com account** credentials. They are only needed when:
- Installing or managing **Splunkbase apps** (`scp_splunkbase_app`)
- Installing **private apps** with pre-vetting (`scp_private_app`)

These credentials authenticate against `splunkbase.splunk.com` (Splunkbase session) and `api.splunk.com` (AppInspect login token). They are needed for Splunkbase app installs, private app pre-vetting workflows, and the `scp_app_validation` data source. If you are only managing indexes, HEC tokens, IP allowlists, users, or roles, you do not need to set these parameters.

## Examples/Documentation 

Refer to the `/examples` directory for example .tf files for each resource and provider configuration. 

Refer to the `/docs` directory for documentation on provider and resource usage, notes, troubleshooting, etc. 


## Contributions 

Currently, we are not accepting contributions, however, please use the
- Github issue tracker to submit bugs
- [Splunk Ideas](https://ideas.splunk.com/) for your suggestions/feature requests. Please file under Enterprise Cloud.  
- [Splunk Answers](https://community.splunk.com/t5/Community/ct-p/en-us) for questions.

## Notes and Troubleshooting 

- If using stack deployment credentials to authenticate, you may run into a rate limit error which prevents the token creation request 
  needed to authenticate. You will need to wait around 5 mins until the request is allowed or use the auth (stack) token to avoid this issue. 




