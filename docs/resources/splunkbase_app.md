# scp_splunkbase_app (Resource in beta version)
### THIS PROVIDER IS AVAILABLE ONLY FOR SPLUNK CLOUD VICTORIA EXPERIENCE
Manages Splunkbase apps in Splunk Cloud Platform. For detailed attribute requirements, refer to the [Splunk Cloud documentation](https://docs.splunk.com/Documentation/SplunkCloud/9.3.2408/Config/ManageSplunkbaseApps) and the ACS API.
Default parallelism for terraform operations is 10. Due to the sequential nature of app installation, it is recommended to use the --parallelism=1 flag when applying Terraform changes with this resource (or at least some number < 10).

This resource supports Targeted Apps Installation (TAI) through the optional `targets` attribute.
When `targets` is set, the provider installs or removes the app on the specified search head targets instead of across the entire stack.

Targeted installs require:

- TAI feature to be enabled on the stack.
- The provider to be configured with `username` and `password` so it can generate per-instance tokens for each target.

If the app was already installed before Terraform began managing targeted installs, you should migrate the resource so Terraform can correctly track the existing target state. The migration process is described below.

## Example Usage

### Global Installation

```terraform
resource "scp_splunkbase_app" "example" {
  name              = "Splunk_TA_nix"
  version           = "8.6.0"
  splunkbase_id     = "833"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
}
```

### Targeted Apps Installation

```terraform
resource "scp_splunkbase_app" "targeted_example" {
  name              = "Splunk_MCP_Server"
  version           = "1.0.5"
  splunkbase_id     = "7931"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
  targets           = ["sh1", "sh2"]
}
```

### Migration to Targeted Installation

If the app was already installed globally before using TAI, follow this process to ensure Terraform tracks all app instances correctly.

**Step 1. Original configuration**

```terraform
resource "scp_splunkbase_app" "targeted_example" {
  name              = "Splunk_MCP_Server"
  version           = "1.0.5"
  splunkbase_id     = "7931"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
}
```

**Step 2. Add all existing targets**

Update the resource to include all search heads where the app is currently installed:

```terraform
resource "scp_splunkbase_app" "targeted_example" {
  name              = "Splunk_MCP_Server"
  version           = "1.0.5"
  splunkbase_id     = "7931"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
  targets           = ["sh1", "sh2"]
}
```

At this point, the app is successfully migrated.

**Step 3. Remove unwanted targets**

You can now remove the app from specific search heads by updating the `targets` list:

```terraform
resource "scp_splunkbase_app" "targeted_example" {
  name              = "Splunk_MCP_Server"
  version           = "1.0.5"
  splunkbase_id     = "7931"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
  targets           = ["sh1"]
}
```

## Schema

### Required

- `name` (`String`): The name of the Splunkbase app.
- `version` (`String`): The version of the Splunkbase app.
- `splunkbase_id` (`String`): The ID of the Splunkbase app.
- `acs_licensing_ack` (`String`): The app's third-party license URL. The license URL is available under `Licensing` on the Splunkbase download page for the app.

### Optional

- `targets` (`Set of String`): The search head targets for Targeted Apps Installation. When set, the provider installs the app only on the specified targets instead of globally across the stack.


## Timeouts
Defaults are currently set to:
- `create` -  30m
- `read` -  30m
- `update` -  30m
- `delete` -  60m
