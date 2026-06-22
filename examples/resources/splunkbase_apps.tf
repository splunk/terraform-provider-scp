
resource "scp_splunkbase_app" "chargeback_app_splunk_cloud" {
  name              = "chargeback_app_splunk_cloud"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
  version           = "2.0.53"
  splunkbase_id     = "5688"
}

resource "scp_splunkbase_app" "broken_hosts" {
  name              = "broken_hosts"
  acs_licensing_ack = "https://opensource.org/licenses/MIT"
  version           = "5.0.4"
  splunkbase_id     = "3247"
}

resource "scp_splunkbase_app" "DomainTools-App-for-Splunk" {
  name              = "DomainTools-App-for-Splunk"
  acs_licensing_ack = "https://cdn.splunkbase.splunk.com/static/misc/eula.html"
  version           = "5.4.0"
  splunkbase_id     = "5226"
}

# Targeted Apps Installation requires TAI to be enabled on the stack and
# provider username/password authentication so per-instance tokens can be generated.
resource "scp_splunkbase_app" "splunk_mcp_server_targeted" {
  name              = "Splunk_MCP_Server"
  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
  version           = "1.0.5"
  splunkbase_id     = "7931"
  targets           = ["sh1", "sh2"]
}
