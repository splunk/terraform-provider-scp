package indexes_test

import (
	"context"
	"fmt"
	"net/http"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
	"github.com/splunk/terraform-provider-scp/internal/indexes"
)

func resourcePrefix(indexName string) string {
	return fmt.Sprint("scp_indexes.", indexName)
}

func TestAcc_SplunkCloudIndex_CreateUpdate(t *testing.T) {
	// Test creating an Index resource and then a new resource with a separate name
	indexCreateResource := resource.UniqueId()
	indexUpdateResource := fmt.Sprintf("%s-%s", indexCreateResource, "new")

	nameResourceTest := []resource.TestStep{
		// Create default index resource
		{
			Config: testAccInstanceConfigBasic(indexCreateResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(indexCreateResource), "name", indexCreateResource),
		},
		{
			Config: testAccInstanceConfigBasic(indexUpdateResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(indexUpdateResource), "name", indexUpdateResource),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps:             nameResourceTest,
	})
}

func TestAcc_SplunkCloudIndex_SearchableDays(t *testing.T) {
	// Test creating an Index resource and then updating searchable_days field
	searchableFieldResource := resource.UniqueId()

	searchableFieldTest := []resource.TestStep{
		// Create default index resource
		{
			Config: testAccInstanceConfigBasic(searchableFieldResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(searchableFieldResource), "name", searchableFieldResource),
		},
		// Update default index with searchable_days field
		{
			Config: testAccInstanceConfigAllFields(searchableFieldResource, "90", "0", "0", ""),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(searchableFieldResource), "searchable_days", "90"),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps:             searchableFieldTest,
	})
}

func TestAcc_SplunkCloudIndex_DatasizeField(t *testing.T) {
	// Test creating an Index resource and then updating max_data_size_mb field
	datasizeFieldResource := resource.UniqueId()

	datasizeFieldTest := []resource.TestStep{
		// Create default index resource
		{
			Config: testAccInstanceConfigBasic(datasizeFieldResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(datasizeFieldResource), "name", datasizeFieldResource),
		},
		// Update default index with max_data_size_mb field
		{
			Config: testAccInstanceConfigAllFields(datasizeFieldResource, "90", "20", "0", ""),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(datasizeFieldResource), "max_data_size_mb", "20"),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps:             datasizeFieldTest,
	})
}

func TestAcc_SplunkCloudIndex_ArchivalField(t *testing.T) {
	// Test creating an Index resource and then updating splunk_archival_retention_days field
	archivalFieldResource := resource.UniqueId()
	regex, err := regexp.Compile("splunkArchivalRetentionDays must be greater than searchableDays")
	if err != nil {
		t.Error()
	}

	archivalFieldTest := []resource.TestStep{
		// Create default index resource
		{
			Config: testAccInstanceConfigBasic(archivalFieldResource),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(archivalFieldResource), "name", archivalFieldResource),
		},
		// Update default index with splunk_archival_retention_days field
		{
			Config: testAccInstanceConfigAllFields(archivalFieldResource, "", "", "120", ""),
			Check:  resource.TestCheckResourceAttr(resourcePrefix(archivalFieldResource), "splunk_archival_retention_days", "120"),
		},
		// Try to update retention days to be less than searchable days expecting failure
		{
			Config:      testAccInstanceConfigAllFields(archivalFieldResource, "", "", "30", ""),
			ExpectError: regex,
			Check:       resource.TestCheckResourceAttr(resourcePrefix(archivalFieldResource), "splunk_archival_retention_days", "120"),
		},
	}

	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps:             archivalFieldTest,
	})
}

// NOTE: The following case can't be automated because it may enter a poll loop due to resource replacement
// Options: remove replacement logic and instead enforce error or skip adding automated tests and inform user of limitation
// Test creating an Index resource and then updating datatype field
//datatypeFieldResource := resource.UniqueId()
//datatypeFieldTest := []resource.TestStep{
//	{
//		Config: testAccInstanceConfigBasic(datatypeFieldResource),
//		Check:  resource.TestCheckResourceAttr(getResourceKey(datatypeFieldResource), "name", datatypeFieldResource),
//	},
//	{
//		Config: testAccInstanceConfigAllFields(datatypeFieldResource, "", "", "", "metric"),
//		Check:  resource.TestCheckResourceAttr(getResourceKey(datatypeFieldResource), "datatype", "metric"),
//	},
//}
//runTerraformTest(t, datatypeFieldTest)

func testAccInstanceConfigBasic(name string) string {
	return fmt.Sprintf("resource \"scp_indexes\" %[1]q {name = %[1]q}", name)
}

func testAccInstanceConfigAllFields(name, searchableDays, maxDataSizeMb, retentionDays, datatype string) string {
	if searchableDays == "" {
		searchableDays = "90"
	}
	if maxDataSizeMb == "" {
		maxDataSizeMb = "0"
	}
	if retentionDays == "" {
		retentionDays = "100"
	}
	if datatype == "" {
		datatype = "event"
	}
	return fmt.Sprintf("resource \"scp_indexes\" %[1]q {\nname = %[1]q \nsearchable_days = %[2]q \nmax_data_size_mb = %[3]q \nsplunk_archival_retention_days = %[4]q \ndatatype = %[5]q \n}", name, searchableDays, maxDataSizeMb, retentionDays, datatype)
}

func testAccCheckIndexDestroy(s *terraform.State) error {
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

		resp, err := acsClient.GetIndexInfo(context.TODO(), stack, v2.Index(rs.Primary.ID))
		if err != nil {
			return fmt.Errorf("Unexpected Error %s", err)
		}
		statusCode := resp.StatusCode
		if statusCode == http.StatusOK {
			return fmt.Errorf("Index still exists")
		} else if statusCode != http.StatusNotFound {
			return fmt.Errorf("expected %d, got %d", 400, statusCode)
		}
	}

	return nil
}
