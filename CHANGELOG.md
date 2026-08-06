# Terraform Provider Splunk Cloud Platform Changelog 

## Version v1.0.0
* Terraform Provider inital MVP release
* Setup provider configuration with token and username/password login support 
* Support for Indexes resource 

## Version v1.1.0
* Terraform Provider Phase Two release 
* Added Indexes datasource 
* Introduced support for HEC Token resource 
* Introduced support for IPAllowlist resource 

## Version v1.2.0
* Terraform Provider Phase Three release
* Introduced support for User resource
* Introduced support for Role resource 
* Added enhancement to HEC token resource to retry previous failed deployment task when creating, updating, deleting Hec Tokens

## Version v1.2.1
* Fixes bug found in Roles resource in which `srch_indexes_default` was set to value of `srch_indexes_allowed`
* Introduces workaround to allow zero value to be set for Roles resource fields where valid. See [Roles Documentation](https://registry.terraform.io/providers/splunk/scp/latest/docs/resources/roles).

## Version v1.2.2
* Adds Support For IPv6 Allowlist Resource
* Go updated to 1.24
* Bugs and Vulnerability Fixes

## Version v1.2.3
* Updates goreleaser flags

## Version v1.2.4
* Hec tokens bug fix to prevent unintended over-writes

## Version v1.2.5
* Index Creation bug fix for index creation timeouts

## Version v1.2.6
* Introduced `_meta` field for HEC token resource

## Version v1.2.7
* Go updated to 1.25.3

## Version v1.3.0
* Support to manage private apps added
* Support to manage Splunkbase apps added

## Version v1.3.1
* Update and reformat documentation for private apps and Splunkbase apps resources

## Version v1.3.2
* Restore: Index Creation bug fix for index creation timeouts

## Version v1.3.3
* Extend Splunkbase apps management to support Targeted app installation

## Version v1.3.4
* Security update: upgrade Go client dependencies to remediate vulnerability findings for `google.golang.org/grpc`, `golang.org/x/net`, and `golang.org/x/crypto`

## Version v1.3.5
* Security update: upgrade `golang.org/x/crypto` and `golang.org/x/net` to remediate vulnerability findings

## Version v1.3.6
* Fixes bug (ACS-146) where setting `force_change_pass = false` on the User resource was not working as intended, leaving the user forced to change their password on first login
