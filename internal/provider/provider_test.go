package provider

import (
	"testing"
)

const VERSION = "1.0.0"

// providerFactories are used to instantiate a provider during acceptance testing.
// The factory function will be invoked for every Terraform CLI command executed
// to create a provider server to which the CLI can reattach.

func TestProvider(t *testing.T) {
	if err := New(VERSION)().InternalValidate(); err != nil {
		t.Fatalf("err: %s", err)
	}
}
