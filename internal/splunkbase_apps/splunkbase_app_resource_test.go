package splunkbaseapps_test

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
	splunkbaseapps "github.com/splunk/terraform-provider-scp/internal/splunkbase_apps"
)

func resourcePrefix(splunkbaseAppName string) string {
	return fmt.Sprint("scp_splunkbase_app.", splunkbaseAppName)
}

func TestAcc_SplunkbaseApps_basic(t *testing.T) {
	appResourceName := "chargeback_app_splunk_cloud"
	splunkbaseID := "5688"
	version := "2.0.52"
	updatedVersion := "2.0.54"
	acsLicensingAck := "https://www.splunk.com/en_us/legal/splunk-general-terms.html"

	updatedFieldsMap := map[string]string{
		"version": "2.0.54",
	}

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfig(appResourceName, version, splunkbaseID, acsLicensingAck),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "name", appResourceName),
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "version", version),
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "splunkbase_id", splunkbaseID),
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "acs_licensing_ack", acsLicensingAck),
				acctest.CheckAppResourceCreated(resourcePrefix(appResourceName)),
			),
		},
		{
			Config: testAccAppConfig(appResourceName, updatedVersion, splunkbaseID, acsLicensingAck),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "name", appResourceName),
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "version", updatedVersion),
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "splunkbase_id", splunkbaseID),
				resource.TestCheckResourceAttr(resourcePrefix(appResourceName), "acs_licensing_ack", acsLicensingAck),
				acctest.CheckAppResourceUpdated(resourcePrefix(appResourceName), updatedFieldsMap),
			),
		},
		{
			Config: testAccAppConfigEmpty(),
			Check: resource.ComposeTestCheckFunc(
				acctest.CheckAppResourceDeleted(resourcePrefix(appResourceName), appResourceName),
			),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkbaseApps_Create(t *testing.T) {
	appCreateResource := "chargeback_app_splunk_cloud"
	splunkbaseID := "5688"
	version := "2.0.53"
	acsLicensingAck := "https://www.splunk.com/en_us/legal/splunk-general-terms.html"

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfig(appCreateResource, version, splunkbaseID, acsLicensingAck),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(appCreateResource), "name", appCreateResource),
		},
		{
			Config: testAccAppConfigMultipleApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "name", "chargeback_app_splunk_cloud"),
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "version", "2.0.53"),
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "splunkbase_id", "5688"),
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "acs_licensing_ack", "https://www.splunk.com/en_us/legal/splunk-general-terms.html"),
				acctest.CheckAppResourceCreated(resourcePrefix("chargeback_app_splunk_cloud")),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "name", "broken_hosts"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "version", "5.0.4"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "splunkbase_id", "3247"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "acs_licensing_ack", "https://opensource.org/licenses/MIT"),
				acctest.CheckAppResourceCreated(resourcePrefix("broken_hosts")),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "name", "DomainTools-App-for-Splunk"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "version", "5.4.1"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "splunkbase_id", "5226"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "acs_licensing_ack", "https://cdn.splunkbase.splunk.com/static/misc/eula.html"),
				acctest.CheckAppResourceCreated(resourcePrefix("DomainTools-App-for-Splunk")),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkbaseApps_Update(t *testing.T) {
	updatedDomainToolsFields := map[string]string{
		"version": "5.5.0",
	}

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfigMultipleApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "name", "broken_hosts"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "version", "5.0.4"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "splunkbase_id", "3247"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "acs_licensing_ack", "https://opensource.org/licenses/MIT"),
				acctest.CheckAppResourceCreated(resourcePrefix("broken_hosts")),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "name", "DomainTools-App-for-Splunk"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "version", "5.4.1"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "splunkbase_id", "5226"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "acs_licensing_ack", "https://cdn.splunkbase.splunk.com/static/misc/eula.html"),
				acctest.CheckAppResourceCreated(resourcePrefix("DomainTools-App-for-Splunk")),
			),
		},
		{
			Config: testAccInstanceConfigUpdatedApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "name", "DomainTools-App-for-Splunk"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "version", "5.5.0"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "splunkbase_id", "5226"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "acs_licensing_ack", "https://cdn.splunkbase.splunk.com/static/misc/eula.html"),
				acctest.CheckAppResourceUpdated(resourcePrefix("DomainTools-App-for-Splunk"), updatedDomainToolsFields),
				acctest.CheckAppResourceDeleted(resourcePrefix("broken_hosts"), "broken_hosts"),
				acctest.CheckAppResourceDeleted(resourcePrefix("chargeback_app_splunk_cloud"), "chargeback_app_splunk_cloud"),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkbaseApps_Delete(t *testing.T) {
	nameResourceTest := []resource.TestStep{
		{
			Config: testAccAppConfigMultipleApps(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "name", "chargeback_app_splunk_cloud"),
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "version", "2.0.53"),
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "splunkbase_id", "5688"),
				resource.TestCheckResourceAttr(resourcePrefix("chargeback_app_splunk_cloud"), "acs_licensing_ack", "https://www.splunk.com/en_us/legal/splunk-general-terms.html"),
				acctest.CheckAppResourceCreated(resourcePrefix("chargeback_app_splunk_cloud")),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "name", "broken_hosts"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "version", "5.0.4"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "splunkbase_id", "3247"),
				resource.TestCheckResourceAttr(resourcePrefix("broken_hosts"), "acs_licensing_ack", "https://opensource.org/licenses/MIT"),
				acctest.CheckAppResourceCreated(resourcePrefix("broken_hosts")),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "name", "DomainTools-App-for-Splunk"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "version", "5.4.1"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "splunkbase_id", "5226"),
				resource.TestCheckResourceAttr(resourcePrefix("DomainTools-App-for-Splunk"), "acs_licensing_ack", "https://cdn.splunkbase.splunk.com/static/misc/eula.html"),
				acctest.CheckAppResourceCreated(resourcePrefix("DomainTools-App-for-Splunk")),
			),
		},
		{
			Config: testAccAppConfigEmpty(),
			Check: resource.ComposeTestCheckFunc(
				acctest.CheckAppResourceDeleted(resourcePrefix("chargeback_app_splunk_cloud"), "chargeback_app_splunk_cloud"),
				acctest.CheckAppResourceDeleted(resourcePrefix("broken_hosts"), "broken_hosts"),
				acctest.CheckAppResourceDeleted(resourcePrefix("DomainTools-App-for-Splunk"), "DomainTools-App-for-Splunk"),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t) },
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
		if rs.Type != splunkbaseapps.ResourceKey {
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
	return "{}"
}

func testAccAppConfig(name, version, splunkbaseID, acsLicensingAck string) string {
	return fmt.Sprintf(`resource "scp_splunkbase_app" %q {
			name = %q
			version = %q
			splunkbase_id = %q
  			acs_licensing_ack = %q
		}`, name, name, version, splunkbaseID, acsLicensingAck)
}

func testAccInstanceConfigUpdatedApps() string {
	return `
		resource "scp_splunkbase_app" "DomainTools-App-for-Splunk" {
		  name             = "DomainTools-App-for-Splunk"
		  acs_licensing_ack = "https://cdn.splunkbase.splunk.com/static/misc/eula.html"
		  version        = "5.5.0"
		  splunkbase_id   = "5226"
		}`
}

func testAccAppConfigMultipleApps() string {
	return `
		resource "scp_splunkbase_app" "chargeback_app_splunk_cloud" {
		  name             = "chargeback_app_splunk_cloud"
		  acs_licensing_ack = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
		  version        = "2.0.53"
		  splunkbase_id   = "5688"
		}
		
		resource "scp_splunkbase_app" "broken_hosts" {
		  name             = "broken_hosts"
		  acs_licensing_ack = "https://opensource.org/licenses/MIT"
		  version        = "5.0.4"
		  splunkbase_id   = "3247"
		}
		
		resource "scp_splunkbase_app" "DomainTools-App-for-Splunk" {
		  name             = "DomainTools-App-for-Splunk"
		  acs_licensing_ack = "https://cdn.splunkbase.splunk.com/static/misc/eula.html"
		  version        = "5.4.1"
		  splunkbase_id   = "5226"
		}`
}
