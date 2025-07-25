package hec_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
	"github.com/splunk/terraform-provider-scp/internal/errors"
	"github.com/splunk/terraform-provider-scp/internal/hec"
)

var (
	defaultIndex        = "main"
	allowedIndexes      = []string{"main", "summary"}
	defaultSourcetype   = "catalina"
	disabled            = "true"
	useAck              = "true"
	invalidDefaultIndex = "invalid"
)

// returns full name of resource with prefix for use in TestCheckResourceAttr
func resourcePrefix(hecName string) string {
	return fmt.Sprint("scp_hec_tokens.", hecName)
}

func TestAcc_SplunkCloudHEC_Create(t *testing.T) {
	// Test creating an HEC resource and then a new resource with a separate name and additional fields
	hecCreateBasicResource := resource.UniqueId()               // bare name of resource
	basicResourceName := resourcePrefix(hecCreateBasicResource) //full name of resource with prefix

	hecCreateNewResource := fmt.Sprintf("%s-%s", hecCreateBasicResource, "new")
	newResourceName := resourcePrefix(hecCreateNewResource)

	hecCreateInvalidResource := fmt.Sprintf("%s-%s", hecCreateBasicResource, "invalid")

	hecCreateEmptyDefaultIndex := fmt.Sprintf("%s-%s", hecCreateBasicResource, "empty-default")

	//Create regexp object for Expect errors
	resourceExistsErr, err := regexp.Compile(errors.ResourceExistsErr)
	if err != nil {
		t.Error()
	}
	acsErr, err := regexp.Compile(errors.AcsErrSuffix)
	if err != nil {
		t.Error()
	}

	emptyDefaultIndexErr, err := regexp.Compile("cannot be an empty string. Either omit it or pick an index from")
	if err != nil {
		t.Error()
	}

	//Generate configs
	basicConfig := testAccInstanceConfigBasic(hecCreateBasicResource)
	testCreateResourceExists := testAccInstanceConfigBasic(hecCreateBasicResource) + testAccInstanceConfigTestCreateResourceExists(hecCreateBasicResource)

	nameResourceTest := []resource.TestStep{
		// Create default hec resource
		{
			Config: basicConfig,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(basicResourceName, "name", hecCreateBasicResource),
				resource.TestCheckResourceAttr(basicResourceName, hec.DefaultIndexKey, "default"),
				resource.TestCheckResourceAttr(basicResourceName, hec.DisabledKey, "false"),
				resource.TestCheckResourceAttr(basicResourceName, hec.UseAckKey, "false"),
			),
		},
		// Create New resource with all fields
		{
			Config: testAccInstanceConfigTestValidCreate(hecCreateNewResource),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(newResourceName, "name", hecCreateNewResource),
				resource.TestCheckResourceAttr(newResourceName, hec.DefaultIndexKey, defaultIndex),
				resource.TestCheckResourceAttr(newResourceName, fmt.Sprint(hec.AllowedIndexesKey, ".", 0), allowedIndexes[0]),
				resource.TestCheckResourceAttr(newResourceName, fmt.Sprint(hec.AllowedIndexesKey, ".", 1), allowedIndexes[1]),
				resource.TestCheckResourceAttr(newResourceName, hec.DefaultSourcetypeKey, defaultSourcetype),
				resource.TestCheckResourceAttr(newResourceName, hec.DisabledKey, disabled),
				resource.TestCheckResourceAttr(newResourceName, hec.UseAckKey, useAck),
			),
		},
		// Expect Error on attempt to create resource that already exists
		{
			Config: testCreateResourceExists,
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(basicResourceName, "name", hecCreateBasicResource),
				resource.TestCheckResourceAttr(basicResourceName, hec.DefaultIndexKey, "main"),
			),
			ExpectError: resourceExistsErr,
		},
		// Expect Error from ACS API
		{
			Config:      testAccInstanceConfigTestAcsErr(hecCreateInvalidResource),
			ExpectError: acsErr,
		},
		// Expect Error from TF is default index is empty string
		{
			Config:      testAccInstanceConfigTestEmptyDefaultIndex(hecCreateEmptyDefaultIndex),
			ExpectError: emptyDefaultIndexErr,
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHecDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkCloudHEC_Update(t *testing.T) {
	// Test update an HEC resource and then a new resource with a separate name
	hecTestUpdateResource := resource.UniqueId()
	resourceName := resourcePrefix(hecTestUpdateResource)

	//Create regexp object for Expect errors
	acsErr, err := regexp.Compile(errors.AcsErrSuffix)
	if err != nil {
		t.Error()
	}

	nameResourceTest := []resource.TestStep{
		// Create default index resource
		{
			Config: testAccInstanceConfigBasic(hecTestUpdateResource),
			Check:  resource.TestCheckResourceAttr(resourceName, "name", hecTestUpdateResource),
		},
		// Update all fields (excluding token and defaultSource)
		{
			Config: testAccInstanceConfigTestValidUpdate(hecTestUpdateResource),
			Check: resource.ComposeTestCheckFunc(
				resource.TestCheckResourceAttr(resourceName, "name", hecTestUpdateResource),
				resource.TestCheckResourceAttr(resourceName, hec.DefaultIndexKey, defaultIndex),
				resource.TestCheckResourceAttr(resourceName, fmt.Sprint(hec.AllowedIndexesKey, ".", 0), allowedIndexes[0]),
				resource.TestCheckResourceAttr(resourceName, fmt.Sprint(hec.AllowedIndexesKey, ".", 1), allowedIndexes[1]),
				resource.TestCheckResourceAttr(resourceName, hec.DefaultSourcetypeKey, defaultSourcetype),
				resource.TestCheckResourceAttr(resourceName, hec.DisabledKey, disabled),
				resource.TestCheckResourceAttr(resourceName, hec.UseAckKey, useAck),
			),
		},
		// Expect Error from ACS API
		{
			Config:      testAccInstanceConfigTestAcsErr(hecTestUpdateResource),
			ExpectError: acsErr,
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckHecDestroy,
		Steps:             nameResourceTest,
	})
}

// Will be run as the last step of each TestCase
func testAccCheckHecDestroy(s *terraform.State) error {
	providerNew := acctest.Provider
	diags := providerNew.Configure(context.Background(), terraform.NewResourceConfigRaw(nil))
	if diags != nil {
		return fmt.Errorf("%+v", diags)

	}
	acsProvider := providerNew.Meta().(client.ACSProvider).Client
	acsClient := *acsProvider
	stack := providerNew.Meta().(client.ACSProvider).Stack

	for _, rs := range s.RootModule().Resources {
		if rs.Type != hec.ResourceKey {
			continue
		}

		resp, err := acsClient.DescribeHec(context.TODO(), stack, v2.Hec(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("Unexpected Error %s", err)
		}
		statusCode := resp.StatusCode
		if statusCode == http.StatusOK {
			return fmt.Errorf("HEC still exists")
		} else if statusCode != http.StatusNotFound {
			return fmt.Errorf("expected %d, got %d", 400, statusCode)
		}
	}

	return nil
}

func testAccInstanceConfigBasic(name string) string {
	return fmt.Sprintf("resource \"scp_hec_tokens\" %[1]q {name = %[1]q}", name)
}

func testAccInstanceConfigTestValidCreate(name string) string {
	allowedIndexesJSON, _ := json.Marshal(allowedIndexes)
	return fmt.Sprintf(`
		resource "scp_hec_tokens" %[1]q {
		name = %[1]q
		default_index = %[2]q
		allowed_indexes = %[3]v
		default_sourcetype = %[4]q
		disabled = %[5]q 
		use_ack = %[6]q
}
`, name, defaultIndex, string(allowedIndexesJSON), defaultSourcetype, disabled, useAck)
}

// Test error on user attempts to create resource that already exists
func testAccInstanceConfigTestCreateResourceExists(name string) string {
	return fmt.Sprintf(`
		resource "scp_hec_tokens" "hec-2" {
			name = %[1]q
		}
`, name)
}

// Test error returned from ACS API
func testAccInstanceConfigTestAcsErr(name string) string {
	allowedIndexesJSON, _ := json.Marshal(allowedIndexes)
	return fmt.Sprintf(`
		resource "scp_hec_tokens" %[1]q {
		name = %[1]q
		default_index = %[2]q
		allowed_indexes = %[3]v
		default_sourcetype = %[4]q
		disabled = %[5]q 
		use_ack = %[6]q
}
`, name, invalidDefaultIndex, string(allowedIndexesJSON), defaultSourcetype, disabled, useAck)
}

// Test error when users does not specify default indexes when allowed indexes is set
func testAccInstanceConfigTestEmptyDefaultIndex(hecName string) string {
	allowedIndexesJSON, _ := json.Marshal(allowedIndexes)
	return fmt.Sprintf(`
	resource "scp_hec_tokens" %[1]q {
		name = %[1]q
		allowed_indexes = %[2]v
		default_index= ""
		default_sourcetype = %[3]q
		disabled = %[4]q 
		use_ack = %[5]q
}`, hecName, string(allowedIndexesJSON), defaultSourcetype, disabled, useAck)
}

// Updates all fields except token, default_source, and default_host
func testAccInstanceConfigTestValidUpdate(name string) string {
	allowedIndexesJSON, _ := json.Marshal(allowedIndexes)
	return fmt.Sprintf(`
		resource "scp_hec_tokens" %[1]q {
		name = %[1]q
		default_index = %[2]q
		allowed_indexes = %[3]v
		default_sourcetype = %[4]q
		disabled = %[5]q 
		use_ack = %[6]q 
}
`, name, defaultIndex, string(allowedIndexesJSON), defaultSourcetype, disabled, useAck)
}

// Test error on user attempts to update Token attribute
// nolint
func testAccInstanceConfigTestInvalidTokenUpdate(name string) string {
	return fmt.Sprintf(`
		resource "scp_hec_tokens" %[1]q {
		name = %[1]q
		token = "some-other-token"
		default_index = "summary"
}
`, name)
}
