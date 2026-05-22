package teamcity

import (
	"context"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &buildConfDataSource{}
	_ datasource.DataSourceWithConfigure = &buildConfDataSource{}
)

func NewBuildConfDataSource() datasource.DataSource {
	return &buildConfDataSource{}
}

type buildConfDataSource struct {
	client *client.Client
}

func (d *buildConfDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *buildConfDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_build_configuration"
}

func (d *buildConfDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "A build configuration is a collection of settings used to start a build and group the sequence of the builds. More info [here](https://www.jetbrains.com/help/teamcity/creating-and-editing-build-configurations.html)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Computed: true,
			},
			"project_id": schema.StringAttribute{
				Computed: true,
			},
			"description": schema.StringAttribute{
				Computed: true,
			},
			"build_type": schema.StringAttribute{
				Computed: true,
			},
			"paused": schema.BoolAttribute{
				Computed: true,
			},
		},
	}
}

func (d *buildConfDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var conf models.BuildTypeDataModel
	diags := req.Config.Get(ctx, &conf)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := d.client.GetBuildType(conf.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading build configuration",
			"Could not read build configuration: "+err.Error(),
		)
		return
	}

	if result == nil {
		resp.Diagnostics.AddError(
			"Build configuration not found",
			"The build configuration with the provided ID does not exist.",
		)
		return
	}

	conf.ID = types.StringValue(result.ID)
	conf.Name = types.StringValue(result.Name)
	conf.ProjectID = types.StringValue(result.GetProjectID())
	conf.Description = types.StringValue(result.Description)
	if result.Type != "" {
		conf.BuildType = types.StringValue(result.Type)
	} else {
		conf.BuildType = types.StringValue("regular")
	}
	conf.Paused = types.BoolValue(result.Paused)

	diags = resp.State.Set(ctx, &conf)
	resp.Diagnostics.Append(diags...)
}
