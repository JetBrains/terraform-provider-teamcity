package teamcity

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource              = &gitResource{}
	_ resource.ResourceWithConfigure = &gitResource{}
)

func NewGitResource() resource.Resource {
	return &gitResource{}
}

type gitResource struct {
	client *client.Client
}

type GitModel struct {
	Url types.String `tfsdk:"url" teamcity:"url"`
}

func (r *gitResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_git"
}

func (r *gitResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"url": {
				Type:     types.StringType,
				Required: true,
			},
		},
	}, nil
}

func (r *gitResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *gitResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan *GitModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.SetParameter(
		"vcs-roots",
		"Root_Reflect",
		"properties/url",
		plan.Url.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}
	plan.Url = types.String{Value: *result}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *gitResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state *GitModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetParameter(
		"vcs-roots",
		"Root_Reflect",
		"properties/url",
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}
	state.Url = types.String{Value: *result}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *gitResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan *GitModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.SetParameter(
		"vcs-roots",
		"Root_Reflect",
		"properties/url",
		plan.Url.Value,
	)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}
	plan.Url = types.String{Value: *result}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *gitResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
