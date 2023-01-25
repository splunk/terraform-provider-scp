package indexes_test

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
	"net/http"
	"testing"
)

const newIndex = `
resource "scp_indexes" "tf-test-index1" {
    name = "tf-test-index1"
}
`

const updateIndex = `
resource "scp_indexes" "tf-test-index1" {
	name = "tf-test-index1"
	searchable_days = 100
}
`

func TestAccSplunkCloudIndex(t *testing.T) {
	resourceName := "scp_indexes.tf-test-index1"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		CheckDestroy:      testAccCheckIndexDestroy,
		Steps: []resource.TestStep{
			{
				Config: newIndex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "searchable_days", "90"),
				),
			},
			{
				Config: updateIndex,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr(resourceName, "searchable_days", "100"),
				),
			},
		},
	})
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
		if rs.Type != "scp_indexes" {
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
