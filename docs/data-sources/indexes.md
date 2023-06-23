# scp_indexes (Data Source)

Index Data Source. Use this data source to reference default indexes (https://docs.splunk.com/Documentation/Splunk/latest/Indexer/Aboutmanagingindexes) or other indexes you do not wish Terraform to execute write operations on.

## Example Usage

```terraform
data "scp_indexes" "main" {
  name = "main"
}
 
data "scp_indexes" "summary" {
  name = "summary"
}
```

## Schema

### Required

- `name` (String) The name of the index.


### Read-Only

- `id` (String) The ID of this resource.


### Note

- If you would like to create, update, or delete an index, please use Index resource (see [Indexes Documentation](../resources/indexes.md) instead. Index Datasource is only used for reading an Index and storing it in `.tfstate`. 



