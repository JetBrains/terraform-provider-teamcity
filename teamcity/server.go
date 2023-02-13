package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ datasource.DataSource              = &serverDataSource{}
	_ datasource.DataSourceWithConfigure = &serverDataSource{}
)

func NewServerDataSource() datasource.DataSource {
	return &serverDataSource{}
}

type serverDataSource struct {
	client *client.Client
}
type serverDataSourceModel struct {
	ID      types.String `tfsdk:"id"`
	Version types.String `tfsdk:"version"`
}

func (d *serverDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *serverDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_server"
}

func (d *serverDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
			},
			"version": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *serverDataSource) Read(ctx context.Context, _ datasource.ReadRequest, resp *datasource.ReadResponse) {
	var state serverDataSourceModel

	version, err := d.client.GetVersion()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to Read version",
			err.Error(),
		)
		return
	}

	state.ID = types.StringValue("placeholder")
	state.Version = types.StringValue(version)

	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}
