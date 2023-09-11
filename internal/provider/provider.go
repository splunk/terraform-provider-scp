package provider

import (
	"context"
	"fmt"
	"github.com/splunk/terraform-provider-scp/internal/hec"
	"github.com/splunk/terraform-provider-scp/internal/users"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/indexes"
	"github.com/splunk/terraform-provider-scp/internal/ipallowlists"
)

func init() {
	// Set descriptions to support markdown syntax, this will be used in document generation
	// and the language server.
	schema.DescriptionKind = schema.StringMarkdown

	// Customize the content of descriptions when output. For example you can add defaults on
	// to the exported descriptions if present.
	// schema.SchemaDescriptionBuilder = func(s *schema.Schema) string {
	// 	desc := s.Description
	// 	if s.Default != nil {
	// 		desc += fmt.Sprintf(" Defaults to `%v`.", s.Default)
	// 	}
	// 	return strings.TrimSpace(desc)
	// }
}

func New(version string) func() *schema.Provider {
	return func() *schema.Provider {
		provider := &schema.Provider{
			Schema:         providerSchema(),
			ResourcesMap:   providerResources(),
			DataSourcesMap: providerDataSources(),
		}

		provider.ConfigureContextFunc = func(ctx context.Context, d *schema.ResourceData) (interface{}, diag.Diagnostics) {
			return configure(ctx, d, version)
		}

		return provider
	}
}

// Returns a map of splunk resources for configuration
func providerResources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		indexes.ResourceKey:      indexes.ResourceIndex(),
		hec.ResourceKey:          hec.ResourceHecToken(),
		ipallowlists.ResourceKey: ipallowlists.ResourceIPAllowlist(),
		users.ResourceKey:        users.ResourceUser(),
	}
}

// Returns a map of Splunk data sources for configuration
func providerDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		indexes.ResourceKey: indexes.DataSourceIndex(),
	}
}

func providerSchema() map[string]*schema.Schema {
	return map[string]*schema.Schema{
		"server": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("ACS_SERVER", nil),
			Description: "ACS API base URL. May also be provided via ACS_SERVER environment variable.",
		},
		"stack": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("SPLUNK_STACK", nil),
			Description: "Stack to perform ACS operations. May also be provided via SPLUNK_STACK environment variable.",
		},
		"auth_token": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			DefaultFunc:  schema.EnvDefaultFunc("STACK_TOKEN", nil),
			AtLeastOneOf: []string{"username"},
			Description: "Authentication tokens, also known as JSON Web Tokens (JWT), are a method for authenticating " +
				"Splunk platform users into the Splunk platform. May also be provided via STACK_TOKEN environment variable.",
		},
		"username": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("STACK_USERNAME", nil),
			Description: "Splunk Cloud Platform deployment username. May also be provided via STACK_USERNAME environment variable.",
		},
		"password": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			RequiredWith: []string{"username"},
			DefaultFunc:  schema.EnvDefaultFunc("STACK_PASSWORD", nil),
			Description:  "Splunk Cloud Platform deployment password. May also be provided via STACK_PASSWORD environment variable.",
		},
	}
}

func configure(ctx context.Context, d *schema.ResourceData, version string) (interface{}, diag.Diagnostics) {
	provider := client.ACSProvider{}

	// initialize stack
	stackName, ok := d.GetOk("stack")
	if !ok || stackName == "" {
		return nil, diag.Errorf("missing Splunk Deployment stack name")
	}
	provider.Stack = v2.Stack(stackName.(string))

	// initialize client to ACS
	server, ok := d.GetOk("server")
	if !ok || server == "" {
		return nil, diag.Errorf("missing server url")
	}

	token, ok := d.GetOk("auth_token")
	if !ok || token == "" {
		tflog.Info(ctx, "No token provided, using stack credentials to generate ephemeral token.")

		username, ok := d.GetOk("username")
		if !ok || username == "" {
			return nil, diag.Errorf("missing Splunk Deployment username, must provide token or stack username/password")
		}

		password, ok := d.GetOk("password")
		if !ok || password == "" {
			return nil, diag.Errorf("missing Splunk Deployment password")
		}

		tmpClient, err := client.GetClientBasicAuth(server.(string), username.(string), password.(string), version)
		if err != nil {
			return nil, diag.FromErr(err)
		}

		token, err = client.GenerateToken(ctx, tmpClient, username.(string), stackName.(string))
		if err != nil {
			return nil, diag.Errorf(fmt.Sprintf("error while generating token: %v", err))
		}
	}

	acsClient, err := client.GetClient(server.(string), token.(string), version)

	if err != nil {
		return nil, diag.FromErr(err)
	}

	provider.Client = &acsClient
	return provider, nil
}
