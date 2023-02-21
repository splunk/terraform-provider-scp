resource "scp_indexes" "index-1" {
  name = "index-1"
  lifecycle {
    prevent_destroy = true
  }
}

resource "scp_indexes" "index-2" {
  name = "index-2"
  searchable_days = 90
  lifecycle {
    prevent_destroy = true
  }
}

resource "scp_indexes" "index-3" {
  name = "index-3"
  searchable_days = 90
  max_data_size_mb = 512
  lifecycle {
    prevent_destroy = true
  }
}
