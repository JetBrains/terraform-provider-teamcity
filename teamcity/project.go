package teamcity

import (
	"context"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &projectResource{}
	_ resource.ResourceWithConfigure = &projectResource{}
)

func NewProjectResource() resource.Resource {
	return &projectResource{}
}

type projectResource struct {
	client *client.Client
}

type projectResourceModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.String `tfsdk:"id"`
}

func (r *projectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

func (r *projectResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"name": {
				Type:     types.StringType,
				Required: true,
			},
			"id": {
				Type:     types.StringType,
				Optional: true,
				Computed: true,
				PlanModifiers: tfsdk.AttributePlanModifiers{
					resource.UseStateForUnknown(),
				},
			},
		},
	}, nil
}

func (r *projectResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *projectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	project := client.Project{
		Name: plan.Name.ValueString(),
	}
	if !plan.Id.IsUnknown() {
		val := plan.Id.ValueString()
		project.Id = &val
	}

	result, err := r.client.NewProject(project)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting project",
			"Cannot set project, unexpected error: "+err.Error(),
		)
		return
	}

	var newState projectResourceModel
	newState.Name = types.StringValue(result.Name)
	newState.Id = types.StringValue(*result.Id)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetProject(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading project",
			"Could not read project settings: "+err.Error(),
		)
		return
	}

	var newState projectResourceModel
	newState.Name = types.StringValue(actual.Name)
	newState.Id = types.StringValue(*actual.Id)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan projectResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState projectResourceModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState projectResourceModel
	resourceId := oldState.Id.ValueString()

	if result, ok := r.setField(resourceId, "name", oldState.Name, plan.Name, &resp.Diagnostics); ok {
		newState.Name = result
	} else {
		return
	}

	if result, ok := r.setField(resourceId, "id", oldState.Id, plan.Id, &resp.Diagnostics); ok {
		newState.Id = result
	} else {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *projectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProject(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting project",
			"Could not delete project, unexpected error: "+err.Error(),
		)
		return
	}
}

func (r *projectResource) setField(id, name string, state, plan types.String, diag *diag.Diagnostics) (types.String, bool) {
	if plan.Equal(state) {
		return state, true
	}

	val := plan.ValueString()

	result, err := r.client.SetField("projects", id, name, &val)
	if err != nil {
		diag.AddError(
			"Error setting project field",
			err.Error(),
		)
		return types.String{}, false
	}

	return types.StringValue(result), true
}
