package main

import (
	"flag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"
	"github.com/splunk/terraform-provider-splunkcloud/internal/provider"
	"github.com/splunk/terraform-provider-splunkcloud/version"
)

func main() {
	var debugMode bool

	flag.BoolVar(&debugMode, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := &plugin.ServeOpts{
		Debug: debugMode,

		ProviderAddr: "registry.terraform.io/splunk/splunkcloud",
		ProviderFunc: provider.New(version.ProviderVersion),
	}

	plugin.Serve(opts)
}
