resource "scp_roles" "role-1" {
  name                          = "role-1"
  capabilities                  = ["accelerate_search"]
  cumulative_rt_srch_jobs_quota = 50
  cumulative_srch_jobs_quota    = 100
  default_app                   = "search"
  imported_roles                = ["user"]
  rt_srch_jobs_quota            = 6
  srch_jobs_quota               = 3
  srch_disk_quota               = 100
  srch_filter                   = "*"
  srch_time_earliest            = -1
  srch_time_win                 = -1
  federated_search_manage_ack   = "Y"
  lifecycle {
    prevent_destroy = true
  }
}