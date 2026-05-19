package splunkbaseapps_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
	"github.com/splunk/terraform-provider-scp/internal/utils"
	splunkbaseapps "github.com/splunk/terraform-provider-scp/internal/splunkbase_apps"
)

const (
	taiAppName         = "chargeback_app_splunk_cloud"
	taiSplunkbaseID    = "5688"
	taiVersion         = "2.0.52"
	taiUpdatedVersion  = "2.0.54"
	taiLicensingAck    = "https://www.splunk.com/en_us/legal/splunk-general-terms.html"
)


func TestAcc_SplunkbaseApps_TAI_CreateWithSingleTarget(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				// Local state checks
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "name", taiAppName),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "splunkbase_id", taiSplunkbaseID),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "acs_licensing_ack", taiLicensingAck),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1}),
				// Remote state checks
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkbaseApps_TAI_CreateWithMultipleTargets(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping multi-target test: TAI_TARGET_2 not set. Multi-target TAI tests require more than 1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargets(target1, target2),
			Check: resource.ComposeTestCheckFunc(
				// Local state checks
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "name", taiAppName),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1, target2}),
				// Remote state checks on both targets
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_AddTarget verifies that a target can be added to an
// existing targeted install (["sh1"] -> ["sh1","sh2"]).
func TestAcc_SplunkbaseApps_TAI_AddTarget(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping add-target test: TAI_TARGET_2 not set. Multi-target TAI tests require more than 1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargets(target1, target2),
			Check: resource.ComposeTestCheckFunc(
				// Local state: now has 2 targets
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				// Remote state: app exists on both targets
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_RemoveTarget verifies that a target can be removed
// from a multi-target install (["sh1","sh2"] -> ["sh1"]).
func TestAcc_SplunkbaseApps_TAI_RemoveTarget(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping remove-target test: TAI_TARGET_2 not set. Multi-target TAI tests require more than 1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargets(target1, target2),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),

				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceDeletedOnTarget(taiAppName, target2),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_ReplaceTargets verifies that targets can be swapped
// (["sh1"] -> ["sh2"]) -- the old target is uninstalled and the new target gets the app.
func TestAcc_SplunkbaseApps_TAI_ReplaceTargets(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping replace-targets test: TAI_TARGET_2 not set. Multi-target TAI tests require more than 1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target2),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),

				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
				acctest.CheckAppResourceDeletedOnTarget(taiAppName, target1),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_UpdateVersionOnTarget verifies that updating the app
// version on a targeted install works correctly.
func TestAcc_SplunkbaseApps_TAI_UpdateVersionOnTarget(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTargetUpdatedVersion(target1),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiUpdatedVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),

				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceUpdatedOnTarget(resourcePrefix(taiAppName), target1, map[string]string{"version": taiUpdatedVersion}),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_DeleteWithTargets verifies that destroying a resource
// with targets correctly uninstalls the app from all targeted search heads.
func TestAcc_SplunkbaseApps_TAI_DeleteWithTargets(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
		{
			Config: testAccAppConfigEmpty(),
			Check: resource.ComposeTestCheckFunc(
				acctest.CheckAppResourceDeletedOnTarget(taiAppName, target1),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_TargetedToGlobal verifies the migration path from a
// targeted install back to a global install (targets removed).
func TestAcc_SplunkbaseApps_TAI_TargetedToGlobal(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigGlobal(),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "0"),
	
				acctest.CheckAppResourceCreated(resourcePrefix(taiAppName)),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_GlobalToTargeted verifies the migration path from a
// global install to a targeted install (targets added to previously global app),
// then uninstalls the app from all targets.
func TestAcc_SplunkbaseApps_TAI_GlobalToTargeted(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping global-to-targeted test: TAI_TARGET_2 not set. Multi-target TAI tests require more than 1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigGlobal(),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "0"),
				acctest.CheckAppResourceCreated(resourcePrefix(taiAppName)),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargets(target1, target2),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1, target2}),

				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
			),
		},
		{
			Config: testAccAppConfigEmpty(),
			Check: resource.ComposeTestCheckFunc(

				acctest.CheckAppResourceDeletedOnTarget(taiAppName, target1),
				acctest.CheckAppResourceDeletedOnTarget(taiAppName, target2),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_VersionUpdateWithMultipleTargets verifies that updating the
// app version propagates to ALL targets when the target list itself is unchanged.
func TestAcc_SplunkbaseApps_TAI_VersionUpdateWithMultipleTargets(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping multi-target version update test: TAI_TARGET_2 not set.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargets(target1, target2),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1, target2}),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargetsUpdatedVersion(target1, target2),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiUpdatedVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1, target2}),

				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
				acctest.CheckAppResourceUpdatedOnTarget(resourcePrefix(taiAppName), target1, map[string]string{"version": taiUpdatedVersion}),
				acctest.CheckAppResourceUpdatedOnTarget(resourcePrefix(taiAppName), target2, map[string]string{"version": taiUpdatedVersion}),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

// TestAcc_SplunkbaseApps_TAI_AddTargetWithVersionUpdate verifies that when both
// the target list and version change simultaneously, the new target gets the new
// version and the kept target is also updated to the new version.
func TestAcc_SplunkbaseApps_TAI_AddTargetWithVersionUpdate(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping add-target-with-version-update test: TAI_TARGET_2 not set.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	nameResourceTest := []resource.TestStep{
		{
			Config: testAccSplunkbaseAppTAIConfigSingleTarget(target1),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "1"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1}),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
			),
		},
		{
			Config: testAccSplunkbaseAppTAIConfigMultipleTargetsUpdatedVersion(target1, target2),
			Check: resource.ComposeTestCheckFunc(

				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "version", taiUpdatedVersion),
				resource.TestCheckResourceAttr(resourcePrefix(taiAppName), "targets.#", "2"),
				acctest.CheckTAITargetsInState(resourcePrefix(taiAppName), []string{target1, target2}),

				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target1),
				acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiAppName), target2),
				acctest.CheckAppResourceUpdatedOnTarget(resourcePrefix(taiAppName), target2, map[string]string{"version": taiUpdatedVersion}),
			),
		},
	}
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t); acctest.PreCheckSplunkbaseApps(t); acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAISplunkbaseAppDestroy,
		Steps:             nameResourceTest,
	})
}

func testAccCheckTAISplunkbaseAppDestroy(s *terraform.State) error {
	providerNew := acctest.Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return fmt.Errorf("%+v", diags)
	}
	acsProvider := providerNew.Meta().(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	for _, rs := range s.RootModule().Resources {
		if rs.Type != splunkbaseapps.ResourceKey {
			continue
		}

		appID := rs.Primary.Attributes["id"]

		resp, err := acsClient.DescribeAppVictoria(context.TODO(), stack, v2.AppName(appID))
		if err != nil {
			return fmt.Errorf("unexpected error: %s", err)
		}
		if resp.StatusCode == http.StatusOK {
			return fmt.Errorf("app %s still exists globally", appID)
		}

		for _, target := range []string{os.Getenv("TAI_TARGET_1"), os.Getenv("TAI_TARGET_2")} {
			if target == "" {
				continue
			}
			targetStack, err := utils.TargetStackName(target, stack)
			if err != nil {
				continue
			}

			targetClient, err := acsProvider.ClientForTarget(context.TODO(), targetStack)
			if err != nil {
				continue
			}

			resp, err := targetClient.DescribeAppVictoria(context.TODO(), targetStack, v2.AppName(appID))
			if err != nil {
				return fmt.Errorf("unexpected error checking target %s: %s", target, err)
			}
			if resp.StatusCode == http.StatusOK {
				return fmt.Errorf("app %s still exists on target %s", appID, target)
			}
		}
	}

	return nil
}


func testAccSplunkbaseAppTAIConfigGlobal() string {
	return fmt.Sprintf(`
	resource "scp_splunkbase_app" %q {
		name              = %q
		splunkbase_id     = %q
		version           = %q
		acs_licensing_ack = %q
	}`, taiAppName, taiAppName, taiSplunkbaseID, taiVersion, taiLicensingAck)
}

func testAccSplunkbaseAppTAIConfigSingleTarget(target string) string {
	return fmt.Sprintf(`
	resource "scp_splunkbase_app" %q {
		name              = %q
		splunkbase_id     = %q
		version           = %q
		acs_licensing_ack = %q
		targets           = [%q]
	}`, taiAppName, taiAppName, taiSplunkbaseID, taiVersion, taiLicensingAck, target)
}

func testAccSplunkbaseAppTAIConfigSingleTargetUpdatedVersion(target string) string {
	return fmt.Sprintf(`
	resource "scp_splunkbase_app" %q {
		name              = %q
		splunkbase_id     = %q
		version           = %q
		acs_licensing_ack = %q
		targets           = [%q]
	}`, taiAppName, taiAppName, taiSplunkbaseID, taiUpdatedVersion, taiLicensingAck, target)
}

func testAccSplunkbaseAppTAIConfigMultipleTargets(target1, target2 string) string {
	return fmt.Sprintf(`
	resource "scp_splunkbase_app" %q {
		name              = %q
		splunkbase_id     = %q
		version           = %q
		acs_licensing_ack = %q
		targets           = [%q, %q]
	}`, taiAppName, taiAppName, taiSplunkbaseID, taiVersion, taiLicensingAck, target1, target2)
}

func testAccSplunkbaseAppTAIConfigMultipleTargetsUpdatedVersion(target1, target2 string) string {
	return fmt.Sprintf(`
	resource "scp_splunkbase_app" %q {
		name              = %q
		splunkbase_id     = %q
		version           = %q
		acs_licensing_ack = %q
		targets           = [%q, %q]
	}`, taiAppName, taiAppName, taiSplunkbaseID, taiUpdatedVersion, taiLicensingAck, target1, target2)
}
