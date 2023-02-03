package teamcity

import (
	"context"
	"reflect"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &vcsRootResource{}
	_ resource.ResourceWithConfigure = &vcsRootResource{}
)

func NewVcsRootResource() resource.Resource {
	return &vcsRootResource{}
}

type vcsRootResource struct {
	client *client.Client
}

type vcsRootResourceModel struct {
	Name      types.String `tfsdk:"name"`
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	//TODO polling interval
	Git *GitPropertiesModel `tfsdk:"git"`
}

type GitPropertiesModel struct {
	//TODO other properties
	Url    types.String `tfsdk:"url" teamcity:"url"`
	Branch types.String `tfsdk:"branch" teamcity:"branch"`
}

func (r *vcsRootResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vcsroot"
}

func (r *vcsRootResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
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
			"project_id": {
				Type:     types.StringType,
				Required: true,
			},
			"git": {
				Required: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"url": {
						Type:     types.StringType,
						Required: true,
					},
					"branch": {
						Type:     types.StringType,
						Required: true,
					},
				}),
			},
		},
	}, nil
}

func (r *vcsRootResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *vcsRootResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vcsRootResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	root := client.VcsRoot{
		Name:    &plan.Name.Value,
		VcsName: "jetbrains.git",
		Project: client.ProjectLocator{
			Id: plan.ProjectId.Value,
		},
		Properties: client.Properties{
			Property: []client.Property{
				{Name: "url", Value: plan.Git.Url.Value},
				{Name: "branch", Value: plan.Git.Branch.Value},
			},
		},
	}

	actual, err := r.client.NewVcsRoot(root)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting VCS root",
			err.Error(),
		)
		return
	}

	var newState vcsRootResourceModel
	err = readState(actual, &newState)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vcsRootResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vcsRootResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetVcsRoot(state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading VCS root",
			err.Error(),
		)
		return
	}

	var newState vcsRootResourceModel
	err = readState(actual, &newState)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func readState(result *client.VcsRoot, state *vcsRootResourceModel) error {
	state.Name = types.String{Value: *result.Name}
	state.Id = types.String{Value: *result.Id}
	state.ProjectId = types.String{Value: result.Project.Id}

	props := make(map[string]string)
	for _, p := range result.Properties.Property {
		props[p.Name] = p.Value
	}
	state.Git = &GitPropertiesModel{
		Url:    types.String{Value: props["url"]},
		Branch: types.String{Value: props["branch"]},
	}

	return nil
}

func (r *vcsRootResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vcsRootResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState vcsRootResourceModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState = vcsRootResourceModel{
		Git: &GitPropertiesModel{},
	}

	refStruct := reflect.TypeOf(plan.Git).Elem()
	fields := reflect.VisibleFields(refStruct)
	for _, field := range fields {
		tfName := field.Name
		restName := field.Tag.Get("teamcity")

		refPlan := reflect.ValueOf(plan.Git).Elem()
		attr := refPlan.FieldByName(tfName).Interface().(types.String)

		result, ok := r.setParameter(plan.Id.Value, "properties/"+restName, attr, &resp.Diagnostics)
		if !ok {
			return
		}

		refNewState := reflect.ValueOf(newState.Git).Elem()
		newAttr := refNewState.FieldByName(tfName).Addr().Interface().(*types.String)
		*newAttr = *result
	}

	if result, ok := r.setParameter(plan.Id.Value, "name", plan.Name, &resp.Diagnostics); ok {
		newState.Name = *result
	} else {
		return
	}

	//TODO update fields
	newState.ProjectId = plan.ProjectId
	newState.Id = plan.Id

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vcsRootResource) setParameter(id string, name string, value types.String, diag *diag.Diagnostics) (*types.String, bool) {
	//TODO check Null
	result, err := r.client.SetParameter("vcs-roots", id, name, value.Value)
	if err != nil {
		diag.AddError(
			"Error setting VCS root field",
			err.Error(),
		)
		return nil, false
	}
	return &types.String{Value: *result}, true
}

func (r *vcsRootResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcsRootResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVcsRoot(state.Id.Value)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting VCS root",
			err.Error(),
		)
		return
	}
}
