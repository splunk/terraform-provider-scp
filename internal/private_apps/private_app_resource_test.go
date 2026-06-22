package privateapps_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
	privateapps "github.com/splunk/terraform-provider-scp/internal/private_apps"
)

func resourcePrefix(PrivateAppName string) string {
	return fmt.Sprint("scp_private_app.", PrivateAppName)
}

func TestAcc_PrivateApps_basic(t *testing.T) {
	updatedFields := map[string]string{
		"filename": "../../examples/test_0-1.1.0.tar.gz",
	}
	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfig(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "name", "test_0"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "filename", "../../examples/test_app.tar.gz"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "acs_legal_ack", "Y"),
				acctest.CheckAppResourceCreated(resourcePrefix("test_0")),
			),
		},
		{
			Config: testAccAppConfigUpdated(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "name", "test_0"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "filename", "../../examples/test_0-1.1.0.tar.gz"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "acs_legal_ack", "Y"),
				acctest.CheckAppResourceUpdated(resourcePrefix("test_0"), updatedFields),
			),
		},
		{
			Config: testAccAppConfigEmpty(),
			Check: resource.ComposeTestCheckFunc(
				acctest.CheckAppResourceDeleted(resourcePrefix("test_0"), "test_0"),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_PrivateApps_Create(t *testing.T) {
	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfigMoreApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "name", "test_0"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "filename", "../../examples/test_app.tar.gz"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "acs_legal_ack", "Y"),
				acctest.CheckAppResourceCreated(resourcePrefix("test_0")),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_PrivateApps_Update(t *testing.T) {
	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfigMoreApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "name", "test_0"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "filename", "../../examples/test_app.tar.gz"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "acs_legal_ack", "Y"),
				acctest.CheckAppResourceCreated(resourcePrefix("test_0")),
			),
		},
		{
			Config: testAccAppConfigMoreUpdatedApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "name", "test_0"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "filename", "../../examples/test_0-1.1.0.tar.gz"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "acs_legal_ack", "Y"),
				acctest.CheckAppResourceCreated(resourcePrefix("test_0")),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_PrivateApps_Delete(t *testing.T) {
	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfigMoreApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "name", "test_0"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "filename", "../../examples/test_app.tar.gz"),
				resource.TestCheckResourceAttr(resourcePrefix("test_0"), "acs_legal_ack", "Y"),
				acctest.CheckAppResourceCreated(resourcePrefix("test_0")),
			),
		},
		{
			Config: testAccAppConfigEmpty(),
			Check: resource.ComposeTestCheckFunc(
				acctest.CheckAppResourceDeleted(resourcePrefix("test_0"), "test_0"),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func testAccCheckAppDestroy(s *terraform.State) error {
	providerNew := acctest.Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return fmt.Errorf("%+v", diags)

	}
	acsProvider := providerNew.Meta().(*client.ACSProvider).Client
	acsClient := *acsProvider
	stack := providerNew.Meta().(*client.ACSProvider).Stack

	for _, rs := range s.RootModule().Resources {
		if rs.Type != privateapps.ResourceKey {
			continue
		}

		resp, err := acsClient.DescribeAppVictoria(context.TODO(), stack, v2.AppName(rs.Primary.Attributes["id"]))
		if err != nil {
			return fmt.Errorf("unexpected Error %s", err)
		}

		statusCode := resp.StatusCode
		if statusCode == http.StatusOK {
			return fmt.Errorf("App still exists")
		} else if statusCode != http.StatusNotFound {
			return fmt.Errorf("expected %d, got %d, %s", http.StatusNotFound, statusCode, resp.Body)
		}
	}

	return nil
}

func testAccAppConfigEmpty() string {
	return `{}`
}

func testAccAppConfig() string {
	return `
	resource "scp_private_app" "test_0" {
  		name = "test_0"
  		filename  = "../../examples/test_app.tar.gz"
  		acs_legal_ack = "Y"	
		pre_vetted = true
	}`
}

func testAccAppConfigUpdated() string {
	return `
	resource "scp_private_app" "test_0" {
  		name = "test_0"
  		filename  = "../../examples/test_0-1.1.0.tar.gz"
  		acs_legal_ack = "Y"	
		pre_vetted = true
	}`
}

func testAccAppConfigMoreApps() string {
	return `
	resource "scp_private_app" "test_0" {
  		name = "test_0"
  		filename  = "../../examples/test_app.tar.gz"
  		acs_legal_ack = "Y"
		pre_vetted = true
	}
		
	resource "scp_private_app" "test_2" {
  		name = "test_2"
  		filename  = "../../examples/test_2.tar.gz"
  		acs_legal_ack = "Y"
		pre_vetted = true
	}`
}

func testAccAppConfigMoreUpdatedApps() string {
	return `
	resource "scp_private_app" "test_0" {
  		name = "test_0"
  		filename  = "../../examples/test_0-1.1.0.tar.gz"
  		acs_legal_ack = "Y"
		pre_vetted = true
	}`
}
