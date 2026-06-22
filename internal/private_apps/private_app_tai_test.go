package privateapps_test

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
	privateapps "github.com/splunk/terraform-provider-scp/internal/private_apps"
	"github.com/splunk/terraform-provider-scp/internal/utils"
)

const (
	taiPrivateAppName   = "test_0"
	taiPrivateAppFile   = "../../examples/test_app.tar.gz"
	taiPrivateAppFileV2 = "../../examples/test_0-1.1.0.tar.gz"
	taiPrivateAcsAck    = "Y"
)


func TestAcc_PrivateApps_TAI_CreateWithSingleTarget(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					// Local state checks
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "name", taiPrivateAppName),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFile),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "acs_legal_ack", taiPrivateAcsAck),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "pre_vetted", "true"),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1}),
					// Remote state check
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
		},
	})
}

func TestAcc_PrivateApps_TAI_CreateWithMultipleTargets(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping multi-target test: TAI_TARGET_2 not set. Requires >1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAIMultipleTargets(target1, target2),
				Check: resource.ComposeTestCheckFunc(
					// Local state checks
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "2"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1, target2}),
					// Remote state checks on both targets
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target2),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_AddTarget verifies that a target can be added to an
// existing targeted install (["sh1"] -> ["sh1","sh2"]).
func TestAcc_PrivateApps_TAI_AddTarget(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping add-target test: TAI_TARGET_2 not set. Requires >1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
			{
				Config: testAccPrivateAppTAIMultipleTargets(target1, target2),
				Check: resource.ComposeTestCheckFunc(
					// Local state: now has 2 targets
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "2"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1, target2}),
					// Remote state: app exists on both targets
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target2),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_RemoveTarget verifies that a target can be removed
// from a multi-target install (["sh1","sh2"] -> ["sh1"]).
func TestAcc_PrivateApps_TAI_RemoveTarget(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping remove-target test: TAI_TARGET_2 not set. Requires >1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAIMultipleTargets(target1, target2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "2"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target2),
				),
			},
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
					acctest.CheckAppResourceDeletedOnTarget(taiPrivateAppName, target2),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_ReplaceTargets verifies that targets can be swapped
// (["sh1"] -> ["sh2"]) -- the old target is uninstalled and the new target gets the app.
func TestAcc_PrivateApps_TAI_ReplaceTargets(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping replace-targets test: TAI_TARGET_2 not set. Requires >1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
			{
				Config: testAccPrivateAppTAISingleTargetOther(target2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target2),
					acctest.CheckAppResourceDeletedOnTarget(taiPrivateAppName, target1),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_DeleteWithTargets verifies that destroying a resource
// with targets correctly uninstalls the app from all targeted search heads.
func TestAcc_PrivateApps_TAI_DeleteWithTargets(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
			{
				Config: testAccAppConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckAppResourceDeletedOnTarget(taiPrivateAppName, target1),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_TargetedToGlobal verifies the migration path from a
// targeted install back to a global install (targets removed).
func TestAcc_PrivateApps_TAI_TargetedToGlobal(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
			{
				Config: testAccPrivateAppTAIGlobal(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "0"),
					acctest.CheckAppResourceCreated(resourcePrefix(taiPrivateAppName)),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_GlobalToTargeted verifies the migration path from a
// global install to a targeted install (targets added to previously global app),
// then uninstalls the app from all targets.
func TestAcc_PrivateApps_TAI_GlobalToTargeted(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping global-to-targeted test: TAI_TARGET_2 not set. Requires >1 search head.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAIGlobal(),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "0"),
					acctest.CheckAppResourceCreated(resourcePrefix(taiPrivateAppName)),
				),
			},
			{
				Config: testAccPrivateAppTAIMultipleTargets(target1, target2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "2"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1, target2}),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target2),
				),
			},
			{
				Config: testAccAppConfigEmpty(),
				Check: resource.ComposeTestCheckFunc(
					acctest.CheckAppResourceDeletedOnTarget(taiPrivateAppName, target1),
					acctest.CheckAppResourceDeletedOnTarget(taiPrivateAppName, target2),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_AddTargetWithFileUpdate verifies that when both
// the target list and file change simultaneously, the new target gets the new
// file and the kept target is also updated to the new file.
func TestAcc_PrivateApps_TAI_AddTargetWithFileUpdate(t *testing.T) {
	if !acctest.HasMultipleTAITargets() {
		t.Skip("Skipping add-target-with-file-update test: TAI_TARGET_2 not set.")
	}

	target1 := acctest.GetTAITarget1()
	target2 := acctest.GetTAITarget2()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFile),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
			{
				Config: testAccPrivateAppTAIMultipleTargetsUpdatedFile(target1, target2),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFileV2),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "2"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1, target2}),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target2),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_FileUpdateOnSameTargets verifies that updating the app
// file (version) on a stable target set works correctly via uninstall+reinstall.
func TestAcc_PrivateApps_TAI_FileUpdateOnSameTargets(t *testing.T) {
	target1 := acctest.GetTAITarget1()

	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAISingleTarget(target1),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFile),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1}),

					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
			{
				Config: testAccPrivateAppTAISingleTargetUpdatedFile(target1),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFileV2),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "1"),
					acctest.CheckTAITargetsInState(resourcePrefix(taiPrivateAppName), []string{target1}),

					acctest.CheckAppResourceCreatedOnTarget(resourcePrefix(taiPrivateAppName), target1),
				),
			},
		},
	})
}

// TestAcc_PrivateApps_TAI_FileUpdateGlobal verifies that updating the app file
// (version) on a global (non-targeted) install works correctly.
func TestAcc_PrivateApps_TAI_FileUpdateGlobal(t *testing.T) {
	resource.Test(t, resource.TestCase{
		PreCheck: func() {
			acctest.PreCheck(t)
			acctest.PreCheckTAI(t)
			acctest.CleanupTAIApp(t, taiPrivateAppName)
		},
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckTAIPrivateAppDestroy,
		Steps: []resource.TestStep{
			{
				Config: testAccPrivateAppTAIGlobal(),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFile),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "0"),

					acctest.CheckAppResourceCreated(resourcePrefix(taiPrivateAppName)),
				),
			},
			{
				Config: testAccPrivateAppTAIGlobalUpdatedFile(),
				Check: resource.ComposeTestCheckFunc(

					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "filename", taiPrivateAppFileV2),
					resource.TestCheckResourceAttr(resourcePrefix(taiPrivateAppName), "targets.#", "0"),

					acctest.CheckAppResourceCreated(resourcePrefix(taiPrivateAppName)),
				),
			},
		},
	})
}

func testAccCheckTAIPrivateAppDestroy(s *terraform.State) error {
	providerNew := acctest.Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return fmt.Errorf("%+v", diags)
	}
	acsProvider := providerNew.Meta().(*client.ACSProvider)
	acsClient := *acsProvider.Client
	stack := acsProvider.Stack

	for _, rs := range s.RootModule().Resources {
		if rs.Type != privateapps.ResourceKey {
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


func testAccPrivateAppTAIGlobal() string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFile, taiPrivateAcsAck)
}

func testAccPrivateAppTAIGlobalUpdatedFile() string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFileV2, taiPrivateAcsAck)
}

func testAccPrivateAppTAISingleTarget(target string) string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
		targets       = [%q]
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFile, taiPrivateAcsAck, target)
}

func testAccPrivateAppTAISingleTargetOther(target string) string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
		targets       = [%q]
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFile, taiPrivateAcsAck, target)
}

func testAccPrivateAppTAISingleTargetUpdatedFile(target string) string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
		targets       = [%q]
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFileV2, taiPrivateAcsAck, target)
}

func testAccPrivateAppTAIMultipleTargets(target1, target2 string) string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
		targets       = [%q, %q]
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFile, taiPrivateAcsAck, target1, target2)
}

func testAccPrivateAppTAIMultipleTargetsUpdatedFile(target1, target2 string) string {
	return fmt.Sprintf(`
	resource "scp_private_app" %q {
		name          = %q
		filename      = %q
		acs_legal_ack = %q
		pre_vetted    = true
		targets       = [%q, %q]
	}`, taiPrivateAppName, taiPrivateAppName, taiPrivateAppFileV2, taiPrivateAcsAck, target1, target2)
}
