# scp_ip_allowlists (Resource)

IP Allowlist Resource. Please see notes to understand unique behavior regarding naming and delete operation.

Please refer to https://docs.splunk.com/Documentation/SplunkCloud/latest/Config/ConfigureIPAllowList 
for more latest, detailed information on attribute requirements and the ACS IP Allowlist API.

## Example Usage

```terraform
resource "scp_ip_allowlists" "hec" {
  feature = "hec"
  subnets = ["###.0.0.0/24", "##.0.10.6/32"]
}
    
resource "scp_ip_allowlists" "search-api" {
  feature = "search-api"
  subnets = ["###.0.0.0/24"] 
}

resource "scp_ip_allowlists" "s2s" {
  feature = "s2s"
  subnets = ["###.0.0.0/24"] 
}
  ```

## Schema

### Required

- `feature` (String) Feature is a specified component in your Splunk Cloud Platform. Eg: search-api, hec, etc. No two 
   resources should have the same feature. Use this value as the resource name itself to enforce this.  
- `subnets` (Set of String) Subnets is a list of IP addresses that have access to the corresponding feature.

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
        subnets = ["###.0.0.0/24", "##.0.10.6/32"]
      }
    
      resource "scp_ip_allowlists" "search-api" {
        feature = "search-api"
        subnets = ["###.0.0.0/24"]
    
      }
      ``` 
    
  - Don't Do: 
      ```terraform
      resource "scp_ip_allowlists" "hec-allowlist-1" {
        feature = "hec"
        subnets = ["###.0.0.0/24"]
      }
    
      resource "scp_ip_allowlists" "hec-allowlist-2" {
        feature = "hec"
        subnets = ["##.0.10.6/32"]
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

