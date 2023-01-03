package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/splunk/terraform-provider-splunkcloud/internal/provider"
)

var (
	// these will be set by the goreleaser configuration to appropriate values for the compiled binary
	version string = "1.0.0"

	// goreleaser can also pass the specific commit if desired
	// commit  string = ""
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug: debugMode,

		ProviderAddr: "registry.terraform.io/splunk/splunkcloud",
		ProviderFunc: provider.New(version),
	}

	plugin.Serve(opts)
}
