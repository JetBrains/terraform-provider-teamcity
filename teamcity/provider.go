package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"os"
	"terraform-provider-teamcity/client"
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

func (p *teamcityProvider) Schema(_ context.Context, _ provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"host": schema.StringAttribute{
				Optional: true,
			},
			"token": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
			"username": schema.StringAttribute{
				Optional: true,
			},
			"password": schema.StringAttribute{
				Optional:  true,
				Sensitive: true,
			},
		},
	}
}

type teamcityProviderModel struct {
	Host     types.String `tfsdk:"host"`
	Token    types.String `tfsdk:"token"`
	Username types.String `tfsdk:"username"`
	Password types.String `tfsdk:"password"`
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
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("username"),
			"Unknown TeamCity username",
			"",
		)
	}
	if config.Token.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("password"),
			"Unknown TeamCity password",
			"",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	host := os.Getenv("TEAMCITY_HOST")
	token := os.Getenv("TEAMCITY_TOKEN")
	username := os.Getenv("TEAMCITY_USERNAME")
	password := os.Getenv("TEAMCITY_PASSWORD")
	if !config.Host.IsNull() {
		host = config.Host.ValueString()
	}
	if !config.Token.IsNull() {
		token = config.Token.ValueString()
	}
	if !config.Username.IsNull() {
		username = config.Username.ValueString()
	}
	if !config.Password.IsNull() {
		password = config.Password.ValueString()
	}

	if host == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("host"),
			"Missing TeamCity Host",
			"",
		)
	}
	if token == "" && username == "" && password == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("token"),
			"Missing TeamCity API Token",
			"",
		)
	}
	if resp.Diagnostics.HasError() {
		return
	}

	cl := client.NewClient(host, token, username, password)
	resp.DataSourceData = &cl
	resp.ResourceData = &cl
}

func (p *teamcityProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewServerDataSource,
		NewBuildConfDataSource,
	}
}

func (p *teamcityProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewCleanupResource,
		NewProjectResource,
		NewSshKeyResource,
		NewVcsRootResource,
		NewVersionedSettingsResource,
		NewRoleResource,
		NewAuthResource,
		NewGlobalResource,
		NewEmailResource,
		NewUserResource,
	}
}
