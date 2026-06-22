resource "scp_private_app" "test" {
  name          = "test_private_app"
  filename      = "../test_app.tar.gz"
  pre_vetted    = true
  acs_legal_ack = "Y"
}

resource "scp_private_app" "test_targeted" {
  name          = "test_private_app"
  filename      = "../test_app.tar.gz"
  pre_vetted    = true
  acs_legal_ack = "Y"
  targets       = ["sh1", "sh2"]
}