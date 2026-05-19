# scp_app_validation (Data Source)

App validation data source. Use this data source to confirm that an existing Splunk AppInspect validation request for a private app has completed successfully before installing that app with `scp_private_app`.

This data source does not submit an app for validation. It reads an existing AppInspect validation request by `request_id`, and the read succeeds only when AppInspect returns a successful result.

## When to use this Data Source

Use `scp_app_validation` when your delivery process already submits an app package to AppInspect outside Terraform and you want Terraform to:

- fail the run if validation did not pass
- pass the resulting `pre_vetted` status into `scp_private_app`

If your organization already tracks vetting outside Terraform and you do not need Terraform to verify the request status, you can set `pre_vetted = true` directly on `scp_private_app` instead.

## Example Usage

### Check an Existing Validation Request

```terraform
data "scp_app_validation" "example" {
  request_id = var.appinspect_request_id
}
```

### Gate a Private App Installation

```terraform
variable "appinspect_request_id" {
  description = "Existing AppInspect validation request ID for this app package"
  type        = string
}

data "scp_app_validation" "private_app" {
  request_id = var.appinspect_request_id
}

resource "scp_private_app" "example" {
  name          = "my-private-app"
  filename      = "/path/to/my-private-app.tar.gz"
  acs_legal_ack = "Y"
  pre_vetted    = data.scp_app_validation.private_app.pre_vetted
}
```

## Schema

### Required

- `request_id` (`String`): The AppInspect validation request ID to read.

### Read-Only

- `id` (`String`): The Terraform state ID for this data source. This is set to the validation request ID.
- `status` (`String`): The current AppInspect status returned for the request.
- `pre_vetted` (`Boolean`): `true` only when validation completed successfully with no errors and no failures.

## Notes

- `request_id` must come from an existing AppInspect submission. This provider does not create validation requests.
- For private app installation details, see [Private App Documentation](../resources/private_app.md).
