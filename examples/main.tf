terraform {
  required_providers {
    scp = {
      source  = "registry.terraform.io/splunk/scp"
    }
  }
}

provider "scp" {
  stack = "example-stack"
  server = "https://admin.splunk.com"
}

resource "scp_indexes" "index-1" {
  name = "index-1"
}

resource "scp_indexes" "index-2" {
  name = "index-2"
  searchable_days = 90
}

resource "scp_indexes" "index-3" {
  name = "index-3"
  searchable_days = 90
  max_data_size_mb = 512
}

resource "scp_indexes" "index-4" {
  name = "index-4"
  searchable_days = 90
  max_data_size_mb = 512
}

data "scp_indexes" "main" {
  name = "main"
}

data "scp_indexes" "history" {
  name = "history"
}