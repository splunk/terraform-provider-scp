terraform {
  required_providers {
    splunkcloud = {
      source  = "registry.terraform.io/splunk/splunkcloud"
    }
  }
}

provider "splunkcloud" {
  stack = "example-stack"
  server = "https://admin.splunk.com"
}

resource "splunkcloud_indexes" "index-1" {
  name = "index-1"
}

resource "splunkcloud_indexes" "index-2" {
  name = "index-2"
  searchable_days = 90
}

resource "splunkcloud_indexes" "index-3" {
  name = "index-3"
  searchable_days = 90
  max_data_size_mb = 512
}

resource "splunkcloud_indexes" "index-4" {
  name = "index-4"
  searchable_days = 90
  max_data_size_mb = 512
}