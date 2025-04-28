package roles_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
	"github.com/splunk/terraform-provider-scp/internal/indexes"
)

var (
	srchJobsQuotaUpdated             = "1"
	srchIndexesAllowed               = []string{}
	srchIndexesAllowedUpdated        = []string{"main"}
	importedRoles                    = []string{"user"}
	cumulativeRTsrchJobsQuota        = "500"
	cumulativeRTsrchJobsQuotaUpdated = "550"
	cumulativeSrchJobsQuota          = "300"
	cumulativeSrchJobsQuotaUpdated   = "350"
	defaultApp                       = "launcher"
	defaultAppUpdated                = "search"
	rtSrchJobsQuota                  = "100"
	rtSrchJobsQuotaUpdated           = "200"
	srchDiskQuota                    = "20"
	srchDiskQuotaUpdated             = "25"
	srchFilter                       = "*"
	srchJobsQuota                    = "100"
	srchTimeEarliest                 = "-1"
	srchTimeWin                      = "-1"
)

func resourcePrefix(roleName string) string {
	return fmt.Sprint("scp_roles.", roleName)
}

func TestAcc_SplunkCloudRole_Create(t *testing.T) {
	// Test creating a role resource and then a new resource with a separate name
	roleCreateResource := resource.UniqueId()
	roleCreateResourceNew := fmt.Sprintf("%s-%s", roleCreateResource, "new")

	nameResourceTest := []resource.TestStep{
		// Create default role resource
		{
			Config: testAccInstanceConfigBasic(roleCreateResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "name", roleCreateResource),
		},
		// Another role
		{
			Config: testAccInstanceConfigBasic(roleCreateResourceNew),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(roleCreateResourceNew), "name", roleCreateResourceNew),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoleDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkCloudRole_UpdateAttributes(t *testing.T) {
	roleCreateResource := resource.UniqueId()

	nameResourceTest := []resource.TestStep{
		// Create default role resource
		{
			Config: testAccInstanceConfigBasic(roleCreateResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "name", roleCreateResource),
		},
		{
			Config: testAccInstanceConfigBasicUpdateAttributes(roleCreateResource),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "cumulative_rt_srch_jobs_quota", cumulativeRTsrchJobsQuotaUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "cumulative_srch_jobs_quota", cumulativeSrchJobsQuotaUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "rt_srch_jobs_quota", rtSrchJobsQuotaUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "srch_jobs_quota", srchJobsQuotaUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "srch_disk_quota", srchDiskQuotaUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "default_app", defaultAppUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "srch_indexes_allowed.#", fmt.Sprint(len(srchIndexesAllowedUpdated))),
				resource.TestCheckResourceAttr(resourcePrefix(roleCreateResource), "srch_indexes_allowed.0", srchIndexesAllowedUpdated[0]),
			),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckRoleDestroy,
		Steps:             nameResourceTest,
	})
}

func testAccInstanceConfigBasic(name string) string {
	roleList, _ := json.Marshal(importedRoles)
	indexList, _ := json.Marshal(srchIndexesAllowed)
	return fmt.Sprintf(`resource "scp_roles" %[1]q {
		name = %[1]q
		cumulative_rt_srch_jobs_quota = %[2]q
		cumulative_srch_jobs_quota = %[3]q
		default_app = %[4]q
		imported_roles = %[5]v
		rt_srch_jobs_quota = %[6]q
		srch_jobs_quota = %[7]q
		srch_disk_quota = %[8]q
		srch_filter = %[9]q
		srch_time_earliest = %[10]q
		srch_time_win = %[11]q
		srch_indexes_allowed = %[12]v
	}`, name, cumulativeRTsrchJobsQuota, cumulativeSrchJobsQuota, defaultApp, string(roleList), rtSrchJobsQuota,
		srchJobsQuota, srchDiskQuota, srchFilter, srchTimeEarliest, srchTimeWin, string(indexList))
}

func testAccInstanceConfigBasicUpdateAttributes(name string) string {
	roleList, _ := json.Marshal(importedRoles)
	updatedIndexList, _ := json.Marshal(srchIndexesAllowedUpdated)
	return fmt.Sprintf(`resource "scp_roles" %[1]q {
		name = %[1]q
		cumulative_rt_srch_jobs_quota = %[2]q
		cumulative_srch_jobs_quota = %[3]q
		default_app = %[4]q
		imported_roles = %[5]v
		rt_srch_jobs_quota = %[6]q
		srch_jobs_quota = %[7]q
		srch_disk_quota = %[8]q
		srch_filter = %[9]q
		srch_time_earliest = %[10]q
		srch_time_win = %[11]q
		srch_indexes_allowed = %[12]v
	}`, name, cumulativeRTsrchJobsQuotaUpdated, cumulativeSrchJobsQuotaUpdated, defaultAppUpdated, string(roleList),
		rtSrchJobsQuotaUpdated, srchJobsQuotaUpdated, srchDiskQuotaUpdated, srchFilter, srchTimeEarliest, srchTimeWin, string(updatedIndexList))
}

func testAccCheckRoleDestroy(s *terraform.State) error {
	providerNew := acctest.Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return fmt.Errorf("%+v", diags)

	}
	acsProvider := providerNew.Meta().(client.ACSProvider).Client
	acsClient := *acsProvider
	stack := providerNew.Meta().(client.ACSProvider).Stack

	for _, rs := range s.RootModule().Resources {
		if rs.Type != indexes.ResourceKey {
			continue
		}

		resp, err := acsClient.DescribeRole(context.TODO(), stack, v2.RoleName(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("Unexpected Error %s", err)
		}
		statusCode := resp.StatusCode
		if statusCode == http.StatusOK {
			return fmt.Errorf("role still exists")
		} else if statusCode != http.StatusNotFound {
			return fmt.Errorf("expected %d, got %d", 400, statusCode)
		}
	}

	return nil
}
