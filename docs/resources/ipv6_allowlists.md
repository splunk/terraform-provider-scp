# scp_ip_v6_allowlists (Resource)

IPv6 Allowlist Resource. Please see notes to understand unique behavior regarding naming and delete operation.

Please refer to https://help.splunk.com/en/splunk-cloud-platform/administer/admin-config-service-manual/9.3.2411/administer-splunk-cloud-platform-using-the-admin-config-service-acs-api/configure-ip-allow-lists-for-splunk-cloud-platform#Configure_IP_allow_lists_for_IPv6
for more latest, detailed information on attribute requirements and the ACS IP Allowlist API.

## Example Usage

```terraform
resource "scp_ip_allowlists" "hec" {
  feature = "hec"
  subnets = ["fe84:1ee:fe23:4637::/64", "2001:db8::ff00:42:8329/128"]
}
    
resource "scp_ip_allowlists" "search-api" {
  feature = "search-api"
  subnets = ["fe84:1ee:fe23:4637::/64"] 
}

resource "scp_ip_allowlists" "s2s" {
  feature = "s2s"
  subnets = ["2001:db8::ff00:42:8329/128"] 
}
  ```

## Schema

### Required

- `feature` (String) Feature is a specified component in your Splunk Cloud Platform. Eg: search-api, hec, etc. No two
  resources should have the same feature. Use this value as the resource name itself to enforce this.
- `subnets` (Set of String) Subnets is a list of IPv6 addresses that have access to the corresponding feature.

### Read-Only

- `id` (String) The ID of this resource.

### NOTE:

- **Must not have two resource blocks where both have the same feature**. Please name the resource in the config file
  the same name as the feature.
  For example:
    - Do
        ```terraform
        resource "scp_ip_allowlists" "hec" {
          feature = "hec"
          subnets = ["fe84:1ee:fe23:4637::/64", "2001:db8::ff00:42:8329/128"]
        }
      
        resource "scp_ip_allowlists" "search-api" {
          feature = "search-api"
          subnets = ["fe84:1ee:fe23:4637::/64"]
      
        }
        ``` 

    - Don't Do:
        ```terraform
        resource "scp_ip_allowlists" "hec-allowlist-1" {
          feature = "hec"
          subnets = ["fe84:1ee:fe23:4637::/64", "2001:db8::ff00:42:8329/128"]
        }
      
        resource "scp_ip_allowlists" "hec-allowlist-2" {
          feature = "hec"
          subnets = ["fe84:1ee:fe23:4637::/64"]
        }
        ``` 
- **IP Allowlist resource does not allow for deletion** of the entire resource. If you
  would like to remove a feature from Terraform management itself, please use:

  ``` terraform state rm scp_indexes.index-1 ```

- Due to API limitations, user **can not update all subnets for a given resource at once**. When updating a subnet list,
  please **keep at least one original subnet** in the list.

## Timeouts
Defaults are currently set to:
- `create` -  20m
- `read` -  20m
- `update` -  20m
- `delete` -  20m

