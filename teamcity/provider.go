package teamcity

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
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
	return tfsdk.Schema{}, nil
}

func (p *teamcityProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
}

func (p *teamcityProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return nil
}

func (p *teamcityProvider) Resources(_ context.Context) []func() resource.Resource {
	return nil
}
