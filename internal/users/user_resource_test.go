package users_test

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
	roles = []string{"user", "sc_admin"}
	// nolint
	password                 = "8bpKOEtnnGVGR2c"
	defaultApp               = "launcher"
	defaultAppUpdated        = "search"
	federatedSearchManageAck = "Y"
	email                    = "tester1@splunk.com"
	emailUpdated             = "tester2@splunk.com"
	fullName                 = "ACS Tester1"
	fullNameUpdated          = "ACS Tester2"
)

func resourcePrefix(userName string) string {
	return fmt.Sprint("scp_users.", userName)
}

func TestAcc_SplunkCloudUser_Create(t *testing.T) {
	// Test creating a user resource and then a new resource with a separate name
	userCreateResource := resource.UniqueId()
	userCreateResourceNew := fmt.Sprintf("%s-%s", userCreateResource, "new")

	nameResourceTest := []resource.TestStep{
		// Create default user resource
		{
			Config: testAccInstanceConfigBasic(userCreateResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(userCreateResource), "name", userCreateResource),
		},
		// Another user
		{
			Config: testAccInstanceConfigBasic(userCreateResourceNew),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(userCreateResourceNew), "name", userCreateResourceNew),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkCloudUser_UpdateAttributes(t *testing.T) {
	userCreateResource := resource.UniqueId()

	nameResourceTest := []resource.TestStep{
		// Create default user resource
		{
			Config: testAccInstanceConfigBasic(userCreateResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(userCreateResource), "name", userCreateResource),
		},
		{
			Config: testAccInstanceConfigBasicUpdateAttributes(userCreateResource),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(userCreateResource), "email", emailUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(userCreateResource), "default_app", defaultAppUpdated),
				resource.TestCheckResourceAttr(resourcePrefix(userCreateResource), "full_name", fullNameUpdated),
			),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckUserDestroy,
		Steps:             nameResourceTest,
	})
}

func testAccInstanceConfigBasic(name string) string {
	roleList, _ := json.Marshal(roles)
	return fmt.Sprintf(`resource "scp_users" %[1]q {
		name = %[1]q
		password = %[2]q
		default_app = %[3]q
		roles = %[4]v
		federated_search_manage_ack = %[5]q
		email = %[6]q
		full_name = %[7]q
	}`, name, password, defaultApp, string(roleList), federatedSearchManageAck, email, fullName)
}

func testAccInstanceConfigBasicUpdateAttributes(name string) string {
	roleList, _ := json.Marshal(roles)
	return fmt.Sprintf(`resource "scp_users" %[1]q {
		name = %[1]q
		password = %[2]q
		default_app = %[3]q
		roles = %[4]v
		federated_search_manage_ack = %[5]q
		email = %[6]q
		full_name = %[7]q
	}`, name, password, defaultAppUpdated, string(roleList), federatedSearchManageAck, emailUpdated, fullNameUpdated)
}

func testAccCheckUserDestroy(s *terraform.State) error {
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

		resp, err := acsClient.DescribeUser(context.TODO(), stack, v2.UserName(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("Unexpected Error %s", err)
		}
		statusCode := resp.StatusCode
		if statusCode == http.StatusOK {
			return fmt.Errorf("user still exists")
		} else if statusCode != http.StatusNotFound {
			return fmt.Errorf("expected %d, got %d", 400, statusCode)
		}
	}

	return nil
}
