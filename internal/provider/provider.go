package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-log/tflog"
	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	v2 "github.com/splunk/terraform-provider-scp/acs/v2"
	"github.com/splunk/terraform-provider-scp/appinspect"
	"github.com/splunk/terraform-provider-scp/client"
	"github.com/splunk/terraform-provider-scp/internal/appvalidation"
	"github.com/splunk/terraform-provider-scp/internal/hec"
	"github.com/splunk/terraform-provider-scp/internal/indexes"
	"github.com/splunk/terraform-provider-scp/internal/ipallowlists"
	"github.com/splunk/terraform-provider-scp/internal/ipv6allowlists"
	privateapps "github.com/splunk/terraform-provider-scp/internal/private_apps"
	"github.com/splunk/terraform-provider-scp/internal/roles"
	splunkbaseapps "github.com/splunk/terraform-provider-scp/internal/splunkbase_apps"
	"github.com/splunk/terraform-provider-scp/internal/users"
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
		indexes.ResourceKey:        indexes.ResourceIndex(),
		hec.ResourceKey:            hec.ResourceHecToken(),
		ipallowlists.ResourceKey:   ipallowlists.ResourceIPAllowlist(),
		ipv6allowlists.ResourceKey: ipv6allowlists.ResourceIPv6Allowlist(),
		roles.ResourceKey:          roles.ResourceRole(),
		users.ResourceKey:          users.ResourceUser(),
		splunkbaseapps.ResourceKey: splunkbaseapps.ResourceSplunkbaseApp(),
		privateapps.ResourceKey:    privateapps.ResourcePrivateApp(),
	}
}

// Returns a map of Splunk data sources for configuration
func providerDataSources() map[string]*schema.Resource {
	return map[string]*schema.Resource{
		indexes.ResourceKey:         indexes.DataSourceIndex(),
		appvalidation.DataSourceKey: appvalidation.DataSourcePrivateAppValidation(),
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
		"splunk_username": {
			Type:        schema.TypeString,
			Optional:    true,
			DefaultFunc: schema.EnvDefaultFunc("SPLUNK_USERNAME", nil),
			Description: "Splunk.com account username. Required only when managing Splunkbase apps, private apps with " +
				"pre-vetting (AppInspect), or the `scp_app_validation` data source. Used to authenticate with " +
				"splunkbase.splunk.com and api.splunk.com. May also be provided via SPLUNK_USERNAME environment variable.",
		},
		"splunk_password": {
			Type:         schema.TypeString,
			Optional:     true,
			Sensitive:    true,
			RequiredWith: []string{"splunk_username"},
			DefaultFunc:  schema.EnvDefaultFunc("SPLUNK_PASSWORD", nil),
			Description: "Splunk.com account password. Required when `splunk_username` is set. Used to authenticate with " +
				"splunkbase.splunk.com and api.splunk.com for Splunkbase app installs and AppInspect validation. " +
				"May also be provided via SPLUNK_PASSWORD environment variable.",
		},
	}
}

func configure(ctx context.Context, d *schema.ResourceData, version string) (interface{}, diag.Diagnostics) {
	provider := client.ACSProvider{}
	stackName, ok := d.GetOk("stack")
	if !ok || stackName == "" {
		return nil, diag.Errorf("missing Splunk Deployment stack name")
	}
	provider.Stack = v2.Stack(stackName.(string))

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
			return nil, diag.Errorf("%s", fmt.Sprintf("error while generating token: %v", err))
		}
	}

	splunkbaseUsername, splunkUsernameOk := d.GetOk("splunk_username")
	splunkbasePassword, splunkPasswordOk := d.GetOk("splunk_password")

	var splunkbaseSession string
	var splunkLoginToken string
	var err error

	if splunkUsernameOk && splunkPasswordOk {
		splunkbaseSession, err = client.GetSplunkbaseSession(ctx, splunkbaseUsername.(string), splunkbasePassword.(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		splunkLoginToken, err = client.GetSplunkLoginToken(splunkbaseUsername.(string), splunkbasePassword.(string))
		if err != nil {
			return nil, diag.FromErr(err)
		}

		appInspectClient := appinspect.GetAppInspectClient(splunkLoginToken)
		provider.AppInspectClient = &appInspectClient
	}

	acsClient, err := client.GetClient(server.(string), token.(string), version, splunkbaseSession, splunkLoginToken)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	provider.Client = &acsClient
	return provider, nil
}
