// Copyright (c) HashiCorp, Inc.
// SPDX-License-Identifier: MPL-2.0

package provider

import (
	"context"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-usgdns/internal/usgdns"
)

// Ensure the implementation satisfies the expected interfaces.
var (
	_ provider.Provider = &usgDnsProvider{}
)

const (
	envCfgUrl   = "USG_DNS_URL"
	envCfgToken = "USG_DNS_TOKEN"
)

type usgDnsProviderModel struct {
	URL   types.String `tfsdk:"url"`
	Token types.String `tfsdk:"token"`
}

// New is a helper function to simplify provider server and testing implementation.
func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &usgDnsProvider{
			version: version,
		}
	}
}

// usgDnsProvider is the provider implementation.
type usgDnsProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// Metadata returns the provider type name.
func (p *usgDnsProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "usgdns"
	resp.Version = p.version
}

// Schema defines the provider-level schema for configuration data.
func (p *usgDnsProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Interact with the usg-dns-api server.",
		Attributes: map[string]schema.Attribute{
			"url": schema.StringAttribute{
				Required:    true,
				Description: "The usg-dns-api server URL. May also be provided via " + envCfgUrl + " environment variable.",
			},
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "The usg-dns-api server token. May also be provided via " + envCfgToken + " environment variable.",
			},
		},
	}
}

// Configure prepares a HashiCups API client for data sources and resources.
func (p *usgDnsProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	// Retrieve provider data from configuration
	var config usgDnsProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// If practitioner provided a configuration value for any of the
	// attributes, it must be a known value.

	if config.URL.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Unknown usg-dns API URL",
			"The provider cannot create the usg-dns API client as there is an unknown configuration value for the URL. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the "+envCfgUrl+" environment variable.",
		)
	}

	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown usg-dns API token",
			"The provider cannot create the usg-dns API client as there is an unknown configuration value for the token. "+
				"Either target apply the source of the value first, set the value statically in the configuration, or use the "+envCfgToken+" environment variable.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Default values to environment variables, but override
	// with Terraform configuration value if set.

	url := os.Getenv(envCfgUrl)
	token := os.Getenv(envCfgToken)

	if !config.URL.IsNull() {
		url = config.URL.ValueString()
	}

	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}

	// If any of the expected configurations are missing, return
	// errors with provider-specific guidance.

	if url == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("url"),
			"Missing usg-dns API URL",
			"The provider cannot create the usg-dns API client as there is a missing or empty value for the URL. "+
				"Set the host value in the configuration or use the "+envCfgUrl+" environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing usg-dns API token",
			"The provider cannot create the usg-dns API client as there is a missing or empty value for the token. "+
				"Set the username value in the configuration or use the "+envCfgToken+" environment variable. "+
				"If either is already set, ensure the value is not empty.",
		)
	}

	if resp.Diagnostics.HasError() {
		return
	}

	// Create a new usg-dns client using the configuration values
	client, err := usgdns.NewClient(url, token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create usg-dns API Client",
			"An unexpected error occurred when creating the usg-dns API client. "+
				"If the error is not clear, please contact the provider developers.\n\n"+
				"usg-dns Client Error: "+err.Error(),
		)
		return
	}

	// Make the usg-dns client available during DataSource and Resource
	// type Configure methods.
	resp.DataSourceData = client
	resp.ResourceData = client
}

// DataSources defines the data sources implemented in the provider.
func (p *usgDnsProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewRecordsDataSource,
	}
}

// Resources defines the resources implemented in the provider.
func (p *usgDnsProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewRecordResource,
	}
}
