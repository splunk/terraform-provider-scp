terraform {
  required_providers {
    scp = {
      source = "registry.terraform.io/splunk/scp"
    }
  }
}

provider "scp" {
  stack = "example-stack"
  server = "https://admin.splunk.com"
  /*****
  To auth to stack use:
  username = "admin" 
  password = var.password
  OR
  auth_token = var.token
  -----------------------
  To auth to splunkbase use:
  splunk_username = var.splunk_username
  splunk_password = var.splunk_password
   *****/
}

resource "scp_hec_tokens" "example-hec-token" {
  name        = "example-hec-token"
  allowed_indexes = ["main"]
  default_index = ["main"]
}

# resource "scp_indexes" "index-1" {
#   name = "index-1"
# }

# resource "scp_indexes" "index-2" {
#   name = "index-2"
#   searchable_days = 90
# }

# resource "scp_indexes" "index-3" {
#   name             = "index-3"
#   searchable_days  = 90
#   max_data_size_mb = 512
# }

# resource "scp_indexes" "index-4" {
#   name             = "index-4"
#   searchable_days  = 90
#   max_data_size_mb = 512
# }

# data "scp_indexes" "main" {
#   name = "main"
# }

# data "scp_indexes" "history" {
#   name = "history"
# }

# resource "scp_private_app" "test_0" {
#   name = "test_0"
#   filename  = "../examples/test_app.tar.gz"
#   acs_legal_ack = "Y"
#   pre_vetted = true
# }
