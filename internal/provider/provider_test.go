package provider

import (
	"github.com/splunk/terraform-provider-splunkcloud/version"
	"testing"
)

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.

func TestProvider(t *testing.T) {
	if err := New(version.ProviderVersion)().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
