package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/splunk/terraform-provider-scp/internal/provider"
	"github.com/splunk/terraform-provider-scp/version"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug: debugMode,

		ProviderAddr: "registry.terraform.io/splunk/scp",
		ProviderFunc: provider.New(version.ProviderVersion),
	}

	plugin.Serve(opts)
}
