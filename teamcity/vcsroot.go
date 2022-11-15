package teamcity

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
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
	Name            types.String        `tfsdk:"name"`
	Id              types.String        `tfsdk:"id"`
	Type            types.String        `tfsdk:"type"`
	PollingInterval types.Int64         `tfsdk:"polling_interval"`
	ProjectId       types.String        `tfsdk:"project_id"`
	Git             *GitPropertiesModel `tfsdk:"git"`
}

type GitPropertiesModel struct {
	Url             types.String `tfsdk:"url"`
	PushUrl         types.String `tfsdk:"push_url"`
	Branch          types.String `tfsdk:"branch"`
	BranchSpec      types.String `tfsdk:"branch_spec"`
	TagsAsBranches  types.Bool   `tfsdk:"tags_as_branches"`
	UsernameStyle   types.String `tfsdk:"username_style"`
	Submodules      types.String `tfsdk:"submodules"`
	UsernameForTags types.String `tfsdk:"username_for_tags"`

	IgnoreKnownHosts types.Bool   `tfsdk:"ignore_known_hosts"`
	ConvertCrlf      types.Bool   `tfsdk:"convert_crlf"`
	PathToGit        types.String `tfsdk:"path_to_git"`
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
			"type": {
				Type:     types.StringType,
				Required: true,
			},
			"polling_interval": {
				Type:     types.Int64Type,
				Optional: true,
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
					"push_url": {
						Type:     types.StringType,
						Optional: true,
					},
					"branch": {
						Type:     types.StringType,
						Required: true,
					},
					"branch_spec": {
						Type:     types.StringType,
						Optional: true,
					},
					"tags_as_branches": {
						Type:     types.BoolType,
						Optional: true,
					},
					"username_style": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf([]string{"USERID", "NAME", "EMAIL", "FULL"}...),
						},
					},
					"submodules": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf([]string{"IGNORE", "CHECKOUT"}...),
						},
					},
					"username_for_tags": {
						Type:     types.StringType,
						Optional: true,
					},
					"ignore_known_hosts": {
						Type:     types.BoolType,
						Optional: true,
					},
					"convert_crlf": {
						Type:     types.BoolType,
						Optional: true,
					},
					"path_to_git": {
						Type:     types.StringType,
						Optional: true,
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

	props := []client.VcsProperty{
		{Name: "url", Value: plan.Git.Url.Value},
		{Name: "branch", Value: plan.Git.Branch.Value},
	}
	if plan.Git.PushUrl.IsNull() != true {
		props = append(props, client.VcsProperty{Name: "push_url", Value: plan.Git.PushUrl.Value})
	}
	if plan.Git.BranchSpec.IsNull() != true {
		props = append(props, client.VcsProperty{Name: "teamcity:branchSpec", Value: plan.Git.BranchSpec.Value})
	}
	if plan.Git.TagsAsBranches.Value == true {
		props = append(props, client.VcsProperty{Name: "reportTagRevisions", Value: "true"})
	} else if plan.Git.TagsAsBranches.Value == false && plan.Git.TagsAsBranches.Null == false {
		props = append(props, client.VcsProperty{Name: "reportTagRevisions", Value: "false"})
	}
	if plan.Git.UsernameStyle.IsNull() != true {
		props = append(props, client.VcsProperty{Name: "usernameStyle", Value: plan.Git.UsernameStyle.Value})
	}
	if plan.Git.Submodules.IsNull() != true {
		props = append(props, client.VcsProperty{Name: "submoduleCheckout", Value: plan.Git.Submodules.Value})
	}
	if plan.Git.UsernameForTags.IsNull() != true {
		props = append(props, client.VcsProperty{Name: "userForTags", Value: plan.Git.UsernameForTags.Value})
	}

	if plan.Git.IgnoreKnownHosts.Value == true {
		props = append(props, client.VcsProperty{Name: "ignoreKnownHosts", Value: "true"})
	} else if plan.Git.IgnoreKnownHosts.Value == false && plan.Git.IgnoreKnownHosts.Null == false {
		props = append(props, client.VcsProperty{Name: "ignoreKnownHosts", Value: "false"})
	}

	if plan.Git.ConvertCrlf.Value == true {
		props = append(props, client.VcsProperty{Name: "serverSideAutoCrlf", Value: "true"})
	} else if plan.Git.ConvertCrlf.Value == false && plan.Git.ConvertCrlf.Null == false {
		props = append(props, client.VcsProperty{Name: "serverSideAutoCrlf", Value: "false"})
	}

	if plan.Git.PathToGit.IsNull() != true {
		props = append(props, client.VcsProperty{Name: "agentGitPath", Value: plan.Git.PathToGit.Value})
	}

	root := client.VcsRoot{
		Name:    &plan.Name.Value,
		VcsName: plan.Type.Value,
		Project: client.ProjectLocator{
			Id: plan.ProjectId.Value,
		},
		Properties: client.VcsProperties{
			Property: props,
		},
	}
	if plan.PollingInterval.IsNull() != true {
		interval := int(plan.PollingInterval.Value)
		root.PollingInterval = &interval
	}

	result, err := r.client.NewVcsRoot(root)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting VCS root",
			err.Error(),
		)
		return
	}

	err = read(result, &plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}

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

	err = read(actual, &state)
	if err != nil {
		resp.Diagnostics.AddError(
			"REST returned invalid value: ",
			err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func read(result *client.VcsRoot, plan *vcsRootResourceModel) error {
	props := make(map[string]string)
	for _, p := range result.Properties.Property {
		props[p.Name] = p.Value
	}

	plan.Name = types.String{Value: *result.Name}
	plan.Id = types.String{Value: *result.Id}

	p := result.PollingInterval
	if p != nil {
		plan.PollingInterval = types.Int64{Value: int64(*result.PollingInterval)}
	} else {
		plan.PollingInterval = types.Int64{Null: true}
	}

	plan.Type = types.String{Value: result.VcsName}
	plan.ProjectId = types.String{Value: result.Project.Id}

	plan.Git = &GitPropertiesModel{
		Url:    types.String{Value: props["url"]},
		Branch: types.String{Value: props["branch"]},
	}

	if val, ok := props["push_url"]; ok {
		plan.Git.PushUrl = types.String{Value: val}
	} else {
		plan.Git.PushUrl = types.String{Null: true}
	}
	if val, ok := props["teamcity:branchSpec"]; ok {
		plan.Git.BranchSpec = types.String{Value: val}
	} else {
		plan.Git.BranchSpec = types.String{Null: true}
	}

	if val, ok := props["reportTagRevisions"]; ok {
		v, err := strconv.ParseBool(val)
		if err == nil {
			plan.Git.TagsAsBranches = types.Bool{Value: v}
		} else {
			return err
		}
	} else {
		plan.Git.TagsAsBranches = types.Bool{Null: true}
	}

	if val, ok := props["usernameStyle"]; ok {
		plan.Git.UsernameStyle = types.String{Value: val}
	} else {
		plan.Git.UsernameStyle = types.String{Null: true}
	}

	if val, ok := props["submoduleCheckout"]; ok {
		plan.Git.Submodules = types.String{Value: val}
	} else {
		plan.Git.Submodules = types.String{Null: true}
	}

	if val, ok := props["userForTags"]; ok {
		plan.Git.UsernameForTags = types.String{Value: val}
	} else {
		plan.Git.UsernameForTags = types.String{Null: true}
	}

	if val, ok := props["ignoreKnownHosts"]; ok {
		v, err := strconv.ParseBool(val)
		if err == nil {
			plan.Git.IgnoreKnownHosts = types.Bool{Value: v}
		} else {
			return err
		}
	} else {
		plan.Git.IgnoreKnownHosts = types.Bool{Null: true}
	}

	if val, ok := props["serverSideAutoCrlf"]; ok {
		v, err := strconv.ParseBool(val)
		if err == nil {
			plan.Git.ConvertCrlf = types.Bool{Value: v}
		} else {
			return err
		}
	} else {
		plan.Git.ConvertCrlf = types.Bool{Null: true}
	}

	if val, ok := props["agentGitPath"]; ok {
		plan.Git.PathToGit = types.String{Value: val}
	} else {
		plan.Git.PathToGit = types.String{Null: true}
	}

	return nil
}

type refType = func(*vcsRootResourceModel) any
type prop struct {
	ref      refType
	resource string
}

func (r *vcsRootResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vcsRootResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var state vcsRootResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	props := []prop{
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.Url },
			resource: "properties/url",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.Branch },
			resource: "properties/branch",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.PushUrl },
			resource: "properties/push_url",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.BranchSpec },
			resource: "properties/teamcity:branchSpec",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.TagsAsBranches },
			resource: "properties/reportTagRevisions",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.UsernameStyle },
			resource: "properties/usernameStyle",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.Submodules },
			resource: "properties/submoduleCheckout",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.UsernameForTags },
			resource: "properties/userForTags",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.IgnoreKnownHosts },
			resource: "properties/ignoreKnownHosts",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.ConvertCrlf },
			resource: "properties/serverSideAutoCrlf",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Git.PathToGit },
			resource: "properties/agentGitPath",
		},
	}

	for _, p := range props {
		err := r.setParameter(&plan, &state, p)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting VCS root field",
				err.Error(),
			)
			return
		}
	}

	fields := []prop{
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.Name },
			resource: "name",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.PollingInterval },
			resource: "modificationCheckInterval",
		},
		{
			ref:      func(a *vcsRootResourceModel) any { return &a.ProjectId },
			resource: "project",
		},
		{ // id is updated last
			ref:      func(a *vcsRootResourceModel) any { return &a.Id },
			resource: "id",
		},
	}

	for _, p := range fields {
		err := r.setParameter(&plan, &state, p)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error setting VCS root field",
				err.Error(),
			)
			return
		}
	}

	if plan.Id.Unknown == true {
		plan.Id = types.String{Value: state.Id.Value}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *vcsRootResource) setParameter(plan, state *vcsRootResourceModel, prop prop) error {
	switch param := prop.ref(plan).(type) {
	case *types.String:
		st := prop.ref(state).(*types.String)
		if param.Unknown != true && param.Value != st.Value {
			result, err := r.client.SetParameter(
				"vcs-roots",
				state.Id.Value,
				prop.resource,
				param.Value,
			)
			if err != nil {
				return err
			}
			param = &types.String{Value: *result}
		}
	case *types.Int64:
		st := prop.ref(state).(*types.Int64)
		if param.Unknown != true && param.Value != st.Value {
			var value string
			if param.IsNull() {
				value = ""
			} else {
				value = param.String()
			}
			result, err := r.client.SetParameter(
				"vcs-roots",
				state.Id.Value,
				prop.resource,
				value,
			)
			if err != nil {
				return err
			}

			i, err := strconv.ParseInt(*result, 10, 64)
			if err != nil {
				return err
			}
			param = &types.Int64{Value: i}
		}
	case *types.Bool:
		st := prop.ref(state).(*types.Bool)
		if param.Unknown == true ||
			st.Null == param.Null && st.Value == param.Value {
		} else {
			var value string
			if param.IsNull() {
				value = ""
			} else if param.Value == false {
				value = "false"
			} else {
				value = "true"
			}

			result, err := r.client.SetParameter(
				"vcs-roots",
				state.Id.Value,
				prop.resource,
				value,
			)
			if err != nil {
				return err
			}

			if *result == "true" {
				param = &types.Bool{Value: true}
			} else if *result == "false " {
				param = &types.Bool{Value: false}
			} else {
				param = &types.Bool{Null: true}
			}
		}
	default:
		return errors.New("Unknown type: " + fmt.Sprintf("%T", param))
	}

	return nil
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
