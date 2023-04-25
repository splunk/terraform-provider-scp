package indexes_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/splunk/terraform-provider-scp/internal/acctest"
)

const indexDataSourceTemplate = `
data "scp_indexes" %[1]q {
	name = %[1]q
}
`

func TestAcc_SplunkCloudIndex_DataSource_basic(t *testing.T) {
	indexName := "main"
	resource.Test(t, resource.TestCase{
		PreCheck:          func() { acctest.PreCheck(t) },
		ProviderFactories: acctest.ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(indexDataSourceTemplate, indexName),
				Check: resource.TestCheckResourceAttr(
					fmt.Sprintf("data.scp_indexes.%s", indexName), "name", indexName,
				),
			},
		},
	})
}
