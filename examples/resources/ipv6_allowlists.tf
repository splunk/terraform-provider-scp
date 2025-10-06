resource "scp_ip_allowlists" "hec" {
  feature = "hec"
  subnets = ["fe84:1ee:fe23:4637::/64", "2001:db8::ff00:42:8329/128"]
}

resource "scp_ipallowlists" "search-api" {
  feature = "search-api"
  subnets = ["fe84:1ee:fe23:4637::/64"]
}

resource "scp_ipallowlists" "s2s" {
  feature = "s2s"
  subnets = ["2001:db8::ff00:42:8329/128"]
}