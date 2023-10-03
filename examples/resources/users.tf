resource "scp_users" "user-1" {
  name                        = "user-1"
  password                    = var.user-1-password
  default_app                 = "launcher"
  roles                       = ["user"]
  federated_search_manage_ack = "Y"
  email                       = "tester@splunk.com"
  full_name                   = "Full Name"
  lifecycle {
    prevent_destroy = true
  }
}