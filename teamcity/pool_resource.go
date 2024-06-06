package teamcity

import (
	"context"
	"errors"
	"fmt"
	"strconv"

	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/setplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"
)

var (
	_ resource.Resource              = &poolResource{}
	_ resource.ResourceWithConfigure = &poolResource{}
)

func NewPoolResource() resource.Resource {
	return &poolResource{}
}

type poolResource struct {
	client *client.Client
}

// Resource functions implementation
// returns the full name of the resource
func (r *poolResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_pool"
}

// returns the schema of the resource
func (r *poolResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "An Agent Pool in TeamCity is a group of agents that can be associated with projects. More info [here](https://www.jetbrains.com/help/teamcity/configuring-agent-pools.html)",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required: true,
			},
			"id": schema.Int64Attribute{
				Computed: true,
				PlanModifiers: []planmodifier.Int64{
					int64planmodifier.UseStateForUnknown(),
				},
			},
			"size": schema.Int64Attribute{
				Required:            false,
				Optional:            true,
				MarkdownDescription: "Agents capacity for the given pool, don't add for unlimited",
				Validators: []validator.Int64{
					int64validator.AtLeast(0),
				},
			},
			"projects": schema.SetAttribute{
				Computed:            true,
				Required:            false,
				Optional:            true,
				MarkdownDescription: "Projects assigned to the given pool",
				ElementType:         types.StringType,
				Default:             setdefault.StaticValue(basetypes.NewSetValueMust(types.StringType, []attr.Value{})),
				PlanModifiers: []planmodifier.Set{
                    setplanmodifier.UseStateForUnknown(),
				},
			},
		},
	}
}

// creates a resource and sets the initial terraform state
func (r *poolResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	// get values from plan
	var plan models.PoolDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Generate API request
	var pool models.PoolJson
	proj := models.ProjectsJson{Project: make([]models.ProjectJson, 0)}
	var size int64

	pool.Name = plan.Name.ValueString()
	if !plan.Size.IsNull() {
		size = plan.Size.ValueInt64()
		pool.Size = &size
	}

	// Create new agent pool
	result, err := r.client.NewPool(pool)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		resp.Diagnostics.AddError(
			"Error creating pool: Timeout",
			err.Error(),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating pool",
			"Cannot create pool, unexpected error: "+err.Error(),
		)
		return
	}

	// Populate computed attributes
	plan.Id = types.Int64Value(int64(*(result.Id)))

	// Two way process: setup associated projects now
	// Assing projects from the plan
    elements := make([]types.String, 0, len(plan.Projects.Elements()))
    diags = plan.Projects.ElementsAs(ctx, &elements, false)
    if diags.HasError() {
        return
    }

	for _, project := range elements {
		id := project.ValueString()
		proj.Project = append(proj.Project, models.ProjectJson{Name: "-", Id: &id})
	}

	response, err := r.client.SetPoolProjects(pool.Name, &proj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting agent pool projects, please check projects IDs are correct",
			err.Error(),
		)
		return
	} else {
        projects := []attr.Value{}
		for _, p := range response.Project {
            projects = append(projects, types.StringValue(string(*p.Id)))
		}

        plan.Projects, diags = types.SetValue(types.StringType, projects)
        if diags.HasError() {
            return
        }
	}

	// Set state
	diags = resp.State.Set(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// reads a resource and sets latest terraform state
func (r *poolResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	// get current state
	var state models.PoolDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get refreshed pool
	pool, err := r.client.GetPool(state.Name.ValueString())
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		resp.Diagnostics.AddError(
			"Agent Pool not found: Timeout",
			err.Error(),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"Agent Pool not found",
			"The Datasource cannot get an Agent Pool since there is no Agent Pool with the provided name.",
		)
		return
	}

	// refresh state to reflect resource not found
	if pool == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	// overwrite with refreshed state
	state = models.PoolDataModel{
		Name: types.StringValue(string(pool.Name)),
		Size: pool.GetSize(),
		Id:   types.Int64Value(int64(*(pool.Id))),
        Projects: types.SetNull(types.StringType),
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

	// set state
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// updates a resource and sets the latest updated terraform state
func (r *poolResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// get values from plan
	var plan models.PoolDataModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// get state
	var state models.PoolDataModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newName string
	var newSize string
	proj := models.ProjectsJson{Project: make([]models.ProjectJson, 0)}

	// Assing projects from the plan
    elements := make([]types.String, 0, len(plan.Projects.Elements()))
    diags = plan.Projects.ElementsAs(ctx, &elements, false)
    if diags.HasError() {
        return
    }

	for _, project := range elements {
		id := project.ValueString()
		proj.Project = append(proj.Project, models.ProjectJson{Name: "-", Id: &id})
	}

	// verify plan values
	newName = plan.Name.ValueString()

	if plan.Size.IsNull() {
		newSize = "-1" // unlimited
	} else {
		size := plan.Size.ValueInt64()
		newSize = strconv.FormatInt(size, 10)
	}

	// verify state id
	if state.Id.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Agent pool state's id cannot be null",
			"The Resource cannot update an Agent Pool since there is an inconsistent state.",
		)
		return
	}

	id := state.Id.String()

	// call update methods
	// Name
	result, err := r.client.SetField("agentPools", id, "name", &newName)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting agent pool name field",
			err.Error(),
		)
		return
	} else {
		state.Name = types.StringValue(result)
	}

	// Size
	result, err = r.client.SetField("agentPools", id, "maxAgents", &newSize)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting agent pool size field",
			err.Error(),
		)
		return
	} else {

		if result == "" {
			state.Size = basetypes.NewInt64Null()
		} else {
			i, err := strconv.ParseInt(result, 10, 64)
			if err != nil {
				resp.Diagnostics.AddError(
					"Could not parse field update response to int64",
					err.Error(),
				)
				return
			}
			state.Size = basetypes.NewInt64Value(i)
		}
	}

	// Projects
	response, err := r.client.SetPoolProjects(newName, &proj)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting agent pool projects, please check projects IDs are correct",
			err.Error(),
		)
		return
	} else {
		projects := []attr.Value{}
        for _, p := range response.Project {
            projects = append(projects, types.StringValue(*p.Id))
        }

        state.Projects, diags = types.SetValue(types.StringType, projects)
        if diags.HasError() {
            return
        }
	}

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

// deletes a resource and removes its terraform state
func (r *poolResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {

	// get state
	var state models.PoolDataModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// verify state id
	if state.Id.IsNull() {
		resp.Diagnostics.AddAttributeError(
			path.Root("id"),
			"Agent pool state's id cannot be null",
			"The Resource cannot delete an Agent Pool since there is an inconsistent state.",
		)
		return
	}

	id := state.Id.String()

	err := r.client.DeletePool(id)
	if err != nil && errors.Is(err, context.DeadlineExceeded) {
		resp.Diagnostics.AddError(
			"Couldn't delete agent pool: Timeout",
			err.Error(),
		)
		return
	}

	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Could not delete pool resource with name %s", state.Name),
			err.Error(),
		)
		return
	}
}

// configure client
func (r *poolResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	client, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	r.client = client
}
