resource "scp_hec_tokens" "hec-1" {
  name = "hec-1"
}

resource "scp_hec_tokens" "hec-2" {
  name = "hec-2"
  allowed_indexes = ["main"]
  use_ack = true
  disabled = false
  lifecycle {
    prevent_destroy = true
  }
}

resource "scp_hec_tokens" "hec-3" {
  name = "hec-3"
  allowed_indexes = ["main", "summary"]
  default_index = "main"
  default_sourcetype = "catalina"
  lifecycle {
    prevent_destroy = true
  }
}