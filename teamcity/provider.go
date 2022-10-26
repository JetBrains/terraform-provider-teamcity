package teamcity

import (
	"context"
	"os"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ provider.Provider = &teamcityProvider{}
)

func New() provider.Provider {
	return &teamcityProvider{}
}

type teamcityProvider struct{}

func (p *teamcityProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "teamcity"
}

func (p *teamcityProvider) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"host": {
				Type:     types.StringType,
				Optional: true,
			},
			"token": {
				Type:      types.StringType,
				Optional:  true,
				Sensitive: true,
			},
		},
	}, nil
}

type teamcityProviderModel struct {
	Host  types.String `tfsdk:"host"`
	Token types.String `tfsdk:"token"`
}

func (p *teamcityProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config teamcityProviderModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Host.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Unknown TeamCity Host",
			"",
		)
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Unknown TeamCity API token",
			"",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("TEAMCITY_HOST")
	token := os.Getenv("TEAMCITY_TOKEN")
	if !config.Host.IsNull() {
		host = config.Host.Value
	}
	if !config.Token.IsNull() {
		token = config.Token.Value
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing TeamCity Host",
			"",
		)
	}
	if token == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing TeamCity API Token",
			"",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	cl, err := client.NewClient(&host, &token)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Create TeamCity Client",
			"Error: "+err.Error(),
		)
		return
	}
	resp.DataSourceData = cl
	resp.ResourceData = cl
}

func (p *teamcityProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerDataSource,
	}
}

func (p *teamcityProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCleanupResource,
		NewProjectResource,
		NewVcsRootResource,
	}
}
