package teamcity

import (
	"context"
	"errors"
	"fmt"

	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
)

var (
	_ datasource.DataSource              = &poolDataSource{}
	_ datasource.DataSourceWithConfigure = &poolDataSource{}
)

func NewPoolDataSource() datasource.DataSource {
	return &poolDataSource{}
}

type poolDataSource struct {
	client *client.Client
}

// DataSource functions implementation
// returns the full name of the data source
func (d *poolDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

// returns the schema of the data source
func (d *poolDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An Agent Pool in TeamCity is a group of agents that can be associated to projects. More info [here](https://www.jetbrains.com/help/teamcity/configuring-agent-pools.html)",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.Int64Attribute{
				Computed: true,
			},
			"size": schema.Int64Attribute{
				Computed:            true,
				MarkdownDescription: "Agents capacity for the given pool",
			},
            "projects": schema.SetAttribute{
                Computed: true,
                ElementType: types.StringType,
            },
		},
	}
}

// returns the state of the data source
func (d *poolDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {

	var name types.String

	diags := req.Config.GetAttribute(ctx, path.Root("name"), &name)
	resp.Diagnostics.Append(diags...)

	if resp.Diagnostics.HasError() {
		return
	}

	if name.IsUnknown() {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Unknown Agent Pool name attribute",
			"The Datasource cannot get an Agent Pool since there is an unknown configuration value for the Agent Pool name.",
		)
		return
	}

	pool, err := d.client.GetPool(name.ValueString())

	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		resp.Diagnostics.AddError(
			"Agent Pool not found: Timeout",
			err.Error(),
		)
		return
	}

	if pool == nil || err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Agent Pool not found",
			"The Datasource cannot get an Agent Pool since there is no Agent Pool with the provided name.",
		)
		return
	}

	var state models.PoolDataModel

	if pool.Size == nil {
		state = models.PoolDataModel{
			Name: types.StringValue(string(pool.Name)),
			Size: basetypes.NewInt64Null(),
			Id:   types.Int64Value(int64(*(pool.Id))),
            Projects: types.SetNull(types.StringType),
		}
	} else {
		state = models.PoolDataModel{
			Name: types.StringValue(string(pool.Name)),
			Size: types.Int64Value(int64(*(pool.Size))),
			Id:   types.Int64Value(int64(*(pool.Id))),
            Projects: types.SetNull(types.StringType),
		}
	}

	if pool.Projects != nil {
        elements := []attr.Value{}
		for _, project := range pool.Projects.Project {
            elements = append(elements, types.StringValue(*project.Id))
		}

        state.Projects, diags = types.SetValue(types.StringType, elements)
        if diags.HasError() {
            return
        }
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// DataSourceWithConfigure functions implementation
func (d *poolDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = client
}
