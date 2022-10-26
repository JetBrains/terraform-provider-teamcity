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
	Type      types.String `tfsdk:"type"`
	ProjectId types.String `tfsdk:"project_id"`
	Url       types.String `tfsdk:"url"`
	Branch    types.String `tfsdk:"branch"`
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
			},
			"type": {
				Type:     types.StringType,
				Required: true,
			},
			"project_id": {
				Type:     types.StringType,
				Required: true,
			},
			"url": {
				Type:     types.StringType,
				Required: true,
			},
			"branch": {
				Type:     types.StringType,
				Required: true,
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
		VcsName: plan.Type.Value,
		Project: client.ProjectLocator{
			Id: plan.ProjectId.Value,
		},
		Properties: client.VcsProperties{
			Property: []client.VcsProperty{
				{Name: "url", Value: plan.Url.Value},
				{Name: "branch", Value: plan.Branch.Value},
			},
		},
	}

	result, err := r.client.NewVcsRoot(root)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting VCS root",
			err.Error(),
		)
		return
	}

	read(result, &plan)

	diags = resp.State.Set(ctx, plan)
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

	read(actual, &state)

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func read(result *client.VcsRoot, plan *vcsRootResourceModel) {
	props := make(map[string]string)
	for _, p := range result.Properties.Property {
		props[p.Name] = p.Value
	}

	plan.Name = types.String{Value: *result.Name}
	plan.Id = types.String{Value: *result.Id}

	plan.Type = types.String{Value: result.VcsName}
	plan.ProjectId = types.String{Value: result.Project.Id}

	plan.Url = types.String{Value: props["url"]}
	plan.Branch = types.String{Value: props["branch"]}
}

func (r *vcsRootResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	//var plan vcsRootResourceModel
	//diags := req.Plan.Get(ctx, &plan)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//var state vcsRootResourceModel
	//diags = req.State.Get(ctx, &state)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
	//
	//result, err := r.client.RenameProject(state.Id.Value, plan.Name.Value)
	//if err != nil {
	//	resp.Diagnostics.AddError(
	//		"Error setting project",
	//		"Cannot set project, unexpected error: "+err.Error(),
	//	)
	//	return
	//}
	//
	//plan.Name = types.String{Value: result.Name}
	//plan.Id = types.String{Value: *result.Id}
	//
	//diags = resp.State.Set(ctx, plan)
	//resp.Diagnostics.Append(diags...)
	//if resp.Diagnostics.HasError() {
	//	return
	//}
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
