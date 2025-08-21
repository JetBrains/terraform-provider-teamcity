package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"terraform-provider-teamcity/client"
)

var (
	_ datasource.DataSource              = &userDataSource{}
	_ datasource.DataSourceWithConfigure = &userDataSource{}
)

type userDataSource struct {
	client *client.Client
}

func NewUserDataSource() datasource.DataSource {
	return &userDataSource{}
}

func (d *userDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_user"
}

type userDataSourceModel struct {
	Id       types.String     `tfsdk:"id"`
	Username types.String     `tfsdk:"username"`
	Github   types.String     `tfsdk:"github_username"`
	Roles    []roleAssignment `tfsdk:"roles"`
}

func (d *userDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Use this data source to get information about an existing TeamCity user.",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The internal ID of the user. Either id or username must be specified.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "The username of the user. Either id or username must be specified.",
			},
			"github_username": schema.StringAttribute{
				Computed:    true,
				Description: "The GitHub username associated with the user.",
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
				Description: "List of roles assigned to the user.",
			},
		},
	}
}

func (d *userDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, _ *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	d.client = req.ProviderData.(*client.Client)
}

func (d *userDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config userDataSourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Validate that either id or username is specified
	if config.Id.IsNull() && config.Username.IsNull() {
		resp.Diagnostics.AddError(
			"Invalid configuration",
			"Either 'id' or 'username' must be specified",
		)
		return
	}

	var user *client.User
	var err error

	// Get user by id or username
	if !config.Id.IsNull() {
		user, err = d.client.GetUser(config.Id.ValueString())
	} else {
		user, err = d.client.GetUserByName(config.Username.ValueString())
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading user",
			err.Error(),
		)
		return
	}

	if user == nil {
		identifier := config.Id.ValueString()
		if config.Id.IsNull() {
			identifier = config.Username.ValueString()
		}
		resp.Diagnostics.AddError(
			"User not found",
			"The user '"+identifier+"' was not found",
		)
		return
	}

	// Map response to model
	state := userDataSourceModel{
		Id:       types.StringValue(strconv.FormatInt(*user.Id, 10)),
		Username: types.StringValue(user.Username),
	}

	// Extract GitHub username from properties
	if user.Properties != nil {
		for _, p := range user.Properties.Property {
			if p.Name == "plugin:auth:GitHubApp-oauth:userName" {
				state.Github = types.StringValue(p.Value)
				break
			}
		}
	}
	if state.Github.IsNull() {
		state.Github = types.StringNull()
	}

	// Map roles
	if user.Roles != nil && len(user.Roles.RoleAssignment) > 0 {
		state.Roles = []roleAssignment{}
		for _, role := range user.Roles.RoleAssignment {
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

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
}
