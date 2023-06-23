package ipallowlists_test

import (
	"fmt"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/splunk/terraform-provider-scp/internal/ipallowlists"
	"github.com/stretchr/testify/assert"
)

/**
DISCLAIMER: NO ACCEPTANCE TEST FOR IP ALLOWLIST RESOURCE

The current implementation of IP Allowlist prevents destroying a resource once created.
This is due to the restrictions placed on the API level to prevent removing all access to a feature.

The existing test framework for TF always attempts to destroy the resource at the end of the test.
`CheckDestroy` function is only triggerred after all test completes, including the deletion of resource.
However, in our case, `CheckDestroy` will never be called since error is thrown when deleting the resource as
part of the Terraform Cleanup.

There have been multiple issues raised on GitHub to support preventing deletion of resource at the end of test.
Until these changes are merged, we are prevented from running Acceptance Test on IP Allowlist feature.
References:
	- https://github.com/hashicorp/terraform-plugin-testing/issues/85
	- https://github.com/hashicorp/terraform-plugin-testing/pull/118

TODO - Action Item once the above changes are merged:
	* Migrate to new testing library: https://github.com/hashicorp/terraform-plugin-testing
	* Refactor existing acceptance test to use this new library
	* Implement acceptance test for IP Allowlist resource on it supports skipping deletion at the end or remove state function
*/

// func TestAcc_SplunkCloudIPAllowlist_AddSubnets(t *testing.T) {
// 	ipAllowlistCreateResource := testFeature
// 	resourcePrefix := getResourcePrefix(t, testFeature)

// 	addSubnetsResourceTest := []resource.TestStep{
// 		{
// 			Config: getIPAllowlistResource(t, ipAllowlistCreateResource, 1),
// 			Check:  resource.TestCheckResourceAttr(resourcePrefix, "feature", ipAllowlistCreateResource),
// 		},
// 	}

// 	resource.Test(t, resource.TestCase{
// 		PreCheck:          func() { acctest.PreCheck(t) },
// 		ProviderFactories: acctest.ProviderFactories,
// 		Steps:             addSubnetsResourceTest,
// 	})
// }

func Test_GetSubnetsFromSet(t *testing.T) {
	t.Run("subnets set is nil", func(t *testing.T) {
		var subnetSet schema.Set
		subnets := ipallowlists.GetSubnetsFromSet(&subnetSet)
		assert.Empty(t, subnets)
	})

	t.Run("subnets is converted correctly", func(t *testing.T) {
		expectedSubnets := getMockSubnets(t, 10)

		subnetsSet := schema.NewSet(schema.HashString, make([]interface{}, 0))
		for _, subnet := range expectedSubnets {
			subnetsSet.Add(subnet)
		}

		gotSubnets := ipallowlists.GetSubnetsFromSet(subnetsSet)
		assert.Equal(t, len(expectedSubnets), len(gotSubnets))
		assert.ElementsMatch(t, expectedSubnets, gotSubnets)
	})
}

// Helper functions
func getMockSubnets(t *testing.T, count int) []string {
	t.Helper()
	subnets := make([]string, count)
	for i := 0; i < count; i++ {
		subnets[i] = fmt.Sprintf("1.1.1.%d/32", (i + 1))
	}
	return subnets
}

// func getIPAllowlistResource(t *testing.T, feature string, subnetCount int) string {
// 	t.Helper()
// 	mockSubnets := getMockSubnets(t, subnetCount)

// 	formattedSubnets := make([]string, len(mockSubnets))
// 	for i, subnet := range mockSubnets {
// 		formattedSubnets[i] = fmt.Sprintf(`"%s"`, subnet)
// 	}

// 	tfSubnets := strings.Join(formattedSubnets, ",")

// 	return fmt.Sprintf(`
//     resource "%[1]s" "%[2]s" {
//         feature = "%[2]s"
//         subnets = [%s]
// 		lifecycle {
// 			prevent_destroy = true
// 		}
//     }
//     `, ipallowlists.ResourceKey, feature, tfSubnets)
// }

// func getResourcePrefix(t *testing.T, feature string) string {
// 	return fmt.Sprintf("%s.%s", ipallowlists.ResourceKey, feature)
// }
