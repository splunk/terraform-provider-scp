package acctest

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/splunk/terraform-provider-scp/internal/provider"
	"github.com/splunk/terraform-provider-scp/version"
	"os"
	"testing"
)

var Provider *schema.Provider

// ProviderFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.
var ProviderFactories = map[string]func() (*schema.Provider, error){
	"scp": func() (*schema.Provider, error) {
		return provider.New(version.ProviderVersion)(), nil
	},
}

func init() {
	var err error
	Provider = provider.New(version.ProviderVersion)()

	if err != nil {
		panic(err)
	}
}

// PreCheck is run prior to any test case execution, add code here to run before any test execution
// For example, assertions about the appropriate environment
func PreCheck(t *testing.T) {
	variables := []string{
		"ACS_SERVER",
		"STACK_TOKEN",
		"SPLUNK_STACK",
		"STACK_USERNAME",
		"STACK_PASSWORD",
	}

	for _, variable := range variables {
		value := os.Getenv(variable)
		if value == "" {
			t.Fatalf("`%s` must be set for acceptance tests!", variable)
		}
	}
}
