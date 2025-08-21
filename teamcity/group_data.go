package teamcity

import (
	"context"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ datasource.DataSource              = &groupDataSource{}
	_ datasource.DataSourceWithConfigure = &groupDataSource{}
)

type groupDataSource struct {
	client *client.Client
}

func NewGroupDataSource() datasource.DataSource {
	return &groupDataSource{}
}

func (d *groupDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_group"
}

type groupDataSourceModel struct {
	Id           types.String     `tfsdk:"id"`
	Key          types.String     `tfsdk:"key"`
	Name         types.String     `tfsdk:"name"`
	Roles        []roleAssignment `tfsdk:"roles"`
	ParentGroups types.Set        `tfsdk:"parent_groups"`
}

func (d *groupDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about an existing TeamCity group.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "The internal ID of the group.",
			},
			"key": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The key (identifier) of the group to retrieve. Either key or name must be specified.",
			},
			"name": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The name of the group to retrieve. Either key or name must be specified.",
			},
			"roles": schema.SetNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"id": schema.StringAttribute{
							Computed:    true,
							Description: "The role ID.",
						},
						"global": schema.BoolAttribute{
							Computed:    true,
							Description: "Whether the role is assigned globally.",
						},
						"project": schema.StringAttribute{
							Computed:    true,
							Description: "The project ID if the role is project-specific.",
						},
					},
				},
				Description: "List of roles assigned to the group.",
			},
			"parent_groups": schema.SetAttribute{
				ElementType: types.StringType,
				Computed:    true,
				Description: "Set of parent group keys.",
			},
		},
	}
}

func (d *groupDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *groupDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config groupDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either key or name is specified
	if config.Key.IsNull() && config.Name.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			"Either 'key' or 'name' must be specified",
		)
		return
	}

	// If both are specified, prefer key
	if !config.Key.IsNull() && !config.Name.IsNull() {
		resp.Diagnostics.AddWarning(
			"Both key and name specified",
			"Both 'key' and 'name' were specified. Using 'key' for the lookup.",
		)
	}

	var group *client.Group
	var err error

	// Get group by key or name
	if !config.Key.IsNull() {
		// Get by key
		group, err = d.client.GetGroup(config.Key.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading group by key",
				err.Error(),
			)
			return
		}
	} else if !config.Name.IsNull() {
		// Get by name - this requires GetGroupByName method in client
		// If the client doesn't have this method, we might need to get all groups and filter
		group, err = d.client.GetGroupByName(config.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading group by name",
				err.Error(),
			)
			return
		}
	}

	if group == nil {
		identifier := config.Key.ValueString()
		if config.Key.IsNull() {
			identifier = "name '" + config.Name.ValueString() + "'"
		} else {
			identifier = "key '" + identifier + "'"
		}
		resp.Diagnostics.AddError(
			"Group not found",
			"The group with "+identifier+" was not found",
		)
		return
	}

	// Map response to model
	state := groupDataSourceModel{
		Id:   types.StringValue(group.Key),
		Key:  types.StringValue(group.Key),
		Name: types.StringValue(group.Name),
	}

	// Map roles
	if group.Roles != nil && len(group.Roles.RoleAssignment) > 0 {
		state.Roles = []roleAssignment{}
		for _, role := range group.Roles.RoleAssignment {
			assignment := roleAssignment{
				Id: types.StringValue(role.Id),
			}
			if role.Scope == "g" {
				assignment.Global = types.BoolValue(true)
			} else {
				assignment.Global = types.BoolValue(false)
				assignment.Project = types.StringValue(role.Scope[2:])
			}
			state.Roles = append(state.Roles, assignment)
		}
	}

	// Map parent groups
	if group.Parents != nil && len(group.Parents.Group) > 0 {
		var parentKeys []attr.Value
		for _, parent := range group.Parents.Group {
			parentKeys = append(parentKeys, types.StringValue(parent.Key))
		}
		state.ParentGroups, _ = types.SetValue(types.StringType, parentKeys)
	} else {
		state.ParentGroups = types.SetNull(types.StringType)
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
