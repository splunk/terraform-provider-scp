resource "scp_ip_allowlists" "hec" {
  feature = "hec"
  subnets = ["###.0.0.0/24", "##.0.10.6/32"]
}

resource "scp_ipallowlists" "search-api" {
  feature = "search-api"
  subnets = ["###.0.0.0/24"]
}

resource "scp_ipallowlists" "s2s" {
  feature = "s2s"
  subnets = ["###.0.0.0/24"]
}