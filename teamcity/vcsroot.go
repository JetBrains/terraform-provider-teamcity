package teamcity

import (
	"context"
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
	Name            types.String `tfsdk:"name"`
	Id              types.String `tfsdk:"id"`
	ProjectId       types.String `tfsdk:"project_id"`
	PollingInterval types.Int64  `tfsdk:"polling_interval"`
	//TODO why pointer?
	Git *GitPropertiesModel `tfsdk:"git"`
}

type GitPropertiesModel struct {
	Url              types.String `tfsdk:"url" teamcity:"url"`
	PushUrl          types.String `tfsdk:"push_url"`
	Branch           types.String `tfsdk:"branch" teamcity:"branch"`
	BranchSpec       types.String `tfsdk:"branch_spec"`
	TagsAsBranches   types.Bool   `tfsdk:"tags_as_branches"`
	UsernameStyle    types.String `tfsdk:"username_style"`
	Submodules       types.String `tfsdk:"submodules"`
	UsernameForTags  types.String `tfsdk:"username_for_tags"`
	AuthMethod       types.String `tfsdk:"auth_method"`
	Username         types.String `tfsdk:"username"`
	Password         types.String `tfsdk:"password"`
	UploadedKey      types.String `tfsdk:"uploaded_key"`
	PrivateKeyPath   types.String `tfsdk:"private_key_path"`
	Passphrase       types.String `tfsdk:"passphrase"`
	IgnoreKnownHosts types.Bool   `tfsdk:"ignore_known_hosts"`
	ConvertCrlf      types.Bool   `tfsdk:"convert_crlf"`
	PathToGit        types.String `tfsdk:"path_to_git"`
	CheckoutPolicy   types.String `tfsdk:"checkout_policy"`
	CleanPolicy      types.String `tfsdk:"clean_policy"`
	CleanFilesPolicy types.String `tfsdk:"clean_files_policy"`
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
			"polling_interval": {
				Type:     types.Int64Type,
				Optional: true,
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
							//TODO other syntax?
							stringvalidator.OneOf([]string{"USERID", "NAME", "EMAIL", "FULL"}...),
						},
					},
					"submodules": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							//TODO other syntax?
							stringvalidator.OneOf([]string{"IGNORE", "CHECKOUT"}...),
						},
					},
					"username_for_tags": {
						Type:     types.StringType,
						Optional: true,
					},
					"auth_method": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf([]string{
								//TODO other syntax? alternate nested types
								"ANONYMOUS",
								"PASSWORD",
								"TEAMCITY_SSH_KEY",
								"PRIVATE_KEY_DEFAULT",
								"PRIVATE_KEY_FILE",
							}...),
						},
					},
					"username": {
						Type:     types.StringType,
						Optional: true,
					},
					"password": {
						Type:      types.StringType,
						Optional:  true,
						Sensitive: true,
					},
					"uploaded_key": {
						Type:     types.StringType,
						Optional: true,
					},
					"private_key_path": {
						Type:     types.StringType,
						Optional: true,
					},
					"passphrase": {
						Type:      types.StringType,
						Optional:  true,
						Sensitive: true,
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
					"checkout_policy": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf([]string{"AUTO", "USE_MIRRORS", "NO_MIRRORS", "SHALLOW_CLONE"}...),
						},
					},
					"clean_policy": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf([]string{"ON_BRANCH_CHANGE", "ALWAYS", "NEVER"}...),
						},
					},
					"clean_files_policy": {
						Type:     types.StringType,
						Optional: true,
						Validators: []tfsdk.AttributeValidator{
							stringvalidator.OneOf([]string{"ALL_UNTRACKED", "IGNORED_ONLY", "NON_IGNORED_ONLY"}...),
						},
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

	var id *string
	if plan.Id.IsUnknown() {
		id = nil
	} else {
		v := plan.Id.ValueString()
		id = &v
	}

	root := client.VcsRoot{
		Name:    plan.Name.ValueString(),
		Id:      id,
		VcsName: "jetbrains.git",
		Project: client.ProjectLocator{
			Id: plan.ProjectId.ValueString(),
		},
	}

	props := []client.Property{
		{Name: "url", Value: plan.Git.Url.ValueString()},
		{Name: "branch", Value: plan.Git.Branch.ValueString()},
	}

	if plan.Git.PushUrl.IsNull() != true {
		props = append(props, client.Property{Name: "push_url", Value: plan.Git.PushUrl.ValueString()})
	}

	if plan.Git.BranchSpec.IsNull() != true {
		props = append(props, client.Property{Name: "teamcity:branchSpec", Value: plan.Git.BranchSpec.ValueString()})
	}

	if plan.Git.TagsAsBranches.IsNull() != true {
		val := strconv.FormatBool(plan.Git.TagsAsBranches.ValueBool())
		props = append(props, client.Property{Name: "reportTagRevisions", Value: val})
	}

	if plan.Git.UsernameStyle.IsNull() != true {
		props = append(props, client.Property{Name: "usernameStyle", Value: plan.Git.UsernameStyle.ValueString()})
	}

	if plan.Git.Submodules.IsNull() != true {
		props = append(props, client.Property{Name: "submoduleCheckout", Value: plan.Git.Submodules.ValueString()})
	}

	if plan.Git.UsernameForTags.IsNull() != true {
		props = append(props, client.Property{Name: "userForTags", Value: plan.Git.UsernameForTags.ValueString()})
	}

	if plan.Git.AuthMethod.IsNull() != true {
		props = append(props, client.Property{Name: "authMethod", Value: plan.Git.AuthMethod.ValueString()})
	}

	if plan.Git.Username.IsNull() != true {
		props = append(props, client.Property{Name: "username", Value: plan.Git.Username.ValueString()})
	}

	if plan.Git.Password.IsNull() != true {
		props = append(props, client.Property{Name: "secure:password", Value: plan.Git.Password.ValueString()})
	}

	if plan.Git.UploadedKey.IsNull() != true {
		props = append(props, client.Property{Name: "teamcitySshKey", Value: plan.Git.UploadedKey.ValueString()})
	}

	if plan.Git.PrivateKeyPath.IsNull() != true {
		props = append(props, client.Property{Name: "privateKeyPath", Value: plan.Git.PrivateKeyPath.ValueString()})
	}

	if plan.Git.Passphrase.IsNull() != true {
		props = append(props, client.Property{Name: "secure:passphrase", Value: plan.Git.Passphrase.ValueString()})
	}

	if plan.Git.IgnoreKnownHosts.IsNull() != true {
		val := strconv.FormatBool(plan.Git.IgnoreKnownHosts.ValueBool())
		props = append(props, client.Property{Name: "ignoreKnownHosts", Value: val})
	}

	if plan.Git.ConvertCrlf.IsNull() != true {
		val := strconv.FormatBool(plan.Git.ConvertCrlf.ValueBool())
		props = append(props, client.Property{Name: "serverSideAutoCrlf", Value: val})
	}

	if plan.Git.PathToGit.IsNull() != true {
		props = append(props, client.Property{Name: "agentGitPath", Value: plan.Git.PathToGit.ValueString()})
	}

	if plan.Git.CheckoutPolicy.IsNull() != true {
		props = append(props, client.Property{Name: "useAlternates", Value: plan.Git.CheckoutPolicy.ValueString()})
	}

	if plan.Git.CleanPolicy.IsNull() != true {
		props = append(props, client.Property{Name: "agentCleanPolicy", Value: plan.Git.CleanPolicy.ValueString()})
	}

	if plan.Git.CleanFilesPolicy.IsNull() != true {
		props = append(props, client.Property{Name: "agentCleanFilesPolicy", Value: plan.Git.CleanFilesPolicy.ValueString()})
	}

	root.Properties = client.Properties{
		Property: props,
	}

	if plan.PollingInterval.IsNull() != true {
		val := int(plan.PollingInterval.ValueInt64())
		root.PollingInterval = &val
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

	actual, err := r.client.GetVcsRoot(state.Id.ValueString())
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

// TODO get rid of pointers: accept object, return new object
func readState(result *client.VcsRoot, state *vcsRootResourceModel) error {
	state.Name = types.StringValue(result.Name)
	state.Id = types.StringValue(*result.Id)
	state.ProjectId = types.StringValue(result.Project.Id)

	if result.PollingInterval != nil {
		state.PollingInterval = types.Int64Value(int64(*result.PollingInterval))
	} else {
		state.PollingInterval = types.Int64Null()
	}

	props := make(map[string]string)
	for _, p := range result.Properties.Property {
		props[p.Name] = p.Value
	}
	state.Git = &GitPropertiesModel{
		Url:    types.StringValue(props["url"]),
		Branch: types.StringValue(props["branch"]),
	}

	if val, ok := props["push_url"]; ok {
		state.Git.PushUrl = types.StringValue(val)
	} else {
		state.Git.PushUrl = types.StringNull()
	}

	if val, ok := props["teamcity:branchSpec"]; ok {
		state.Git.BranchSpec = types.StringValue(val)
	} else {
		state.Git.BranchSpec = types.StringNull()
	}

	if val, ok := props["reportTagRevisions"]; ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		state.Git.TagsAsBranches = types.BoolValue(v)
	} else {
		state.Git.TagsAsBranches = types.BoolNull()
	}

	if val, ok := props["usernameStyle"]; ok {
		state.Git.UsernameStyle = types.StringValue(val)
	} else {
		state.Git.UsernameStyle = types.StringNull()
	}

	if val, ok := props["submoduleCheckout"]; ok {
		state.Git.Submodules = types.StringValue(val)
	} else {
		state.Git.Submodules = types.StringNull()
	}

	if val, ok := props["userForTags"]; ok {
		state.Git.UsernameForTags = types.StringValue(val)
	} else {
		state.Git.UsernameForTags = types.StringNull()
	}

	if val, ok := props["authMethod"]; ok {
		state.Git.AuthMethod = types.StringValue(val)
	} else {
		state.Git.AuthMethod = types.StringNull()
	}

	if val, ok := props["username"]; ok {
		state.Git.Username = types.StringValue(val)
	} else {
		state.Git.Username = types.StringNull()
	}

	if val, ok := props["secure:password"]; ok {
		state.Git.Password = types.StringValue(val)
	} else {
		state.Git.Password = types.StringNull()
	}

	if val, ok := props["teamcitySshKey"]; ok {
		state.Git.UploadedKey = types.StringValue(val)
	} else {
		state.Git.UploadedKey = types.StringNull()
	}

	if val, ok := props["privateKeyPath"]; ok {
		state.Git.PrivateKeyPath = types.StringValue(val)
	} else {
		state.Git.PrivateKeyPath = types.StringNull()
	}

	if val, ok := props["secure:passphrase"]; ok {
		state.Git.Passphrase = types.StringValue(val)
	} else {
		state.Git.Passphrase = types.StringNull()
	}

	if val, ok := props["ignoreKnownHosts"]; ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		state.Git.IgnoreKnownHosts = types.BoolValue(v)
	} else {
		state.Git.IgnoreKnownHosts = types.BoolNull()
	}

	if val, ok := props["serverSideAutoCrlf"]; ok {
		v, err := strconv.ParseBool(val)
		if err != nil {
			return err
		}
		state.Git.ConvertCrlf = types.BoolValue(v)
	} else {
		state.Git.ConvertCrlf = types.BoolNull()
	}

	if val, ok := props["agentGitPath"]; ok {
		state.Git.PathToGit = types.StringValue(val)
	} else {
		state.Git.PathToGit = types.StringNull()
	}

	if val, ok := props["useAlternates"]; ok {
		state.Git.CheckoutPolicy = types.StringValue(val)
	} else {
		state.Git.CheckoutPolicy = types.StringNull()
	}

	if val, ok := props["agentCleanPolicy"]; ok {
		state.Git.CleanPolicy = types.StringValue(val)
	} else {
		state.Git.CleanPolicy = types.StringNull()
	}

	if val, ok := props["agentCleanFilesPolicy"]; ok {
		state.Git.CleanFilesPolicy = types.StringValue(val)
	} else {
		state.Git.CleanFilesPolicy = types.StringNull()
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

	resourceId := oldState.Id.ValueString()

	if result, ok := r.setFieldString(resourceId, "name", oldState.Name, plan.Name, &resp.Diagnostics); ok {
		newState.Name = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "project", oldState.ProjectId, plan.ProjectId, &resp.Diagnostics); ok {
		newState.ProjectId = result
	} else {
		return
	}

	if result, ok := r.setFieldInt(resourceId, "modificationCheckInterval", oldState.PollingInterval, plan.PollingInterval, &resp.Diagnostics); ok {
		newState.PollingInterval = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/url", oldState.Git.Url, plan.Git.Url, &resp.Diagnostics); ok {
		newState.Git.Url = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/push_url", oldState.Git.PushUrl, plan.Git.PushUrl, &resp.Diagnostics); ok {
		newState.Git.PushUrl = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/branch", oldState.Git.Branch, plan.Git.Branch, &resp.Diagnostics); ok {
		newState.Git.Branch = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/teamcity:branchSpec", oldState.Git.BranchSpec, plan.Git.BranchSpec, &resp.Diagnostics); ok {
		newState.Git.BranchSpec = result
	} else {
		return
	}

	if result, ok := r.setFieldBool(resourceId, "properties/reportTagRevisions", oldState.Git.TagsAsBranches, plan.Git.TagsAsBranches, &resp.Diagnostics); ok {
		newState.Git.TagsAsBranches = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/usernameStyle", oldState.Git.UsernameStyle, plan.Git.UsernameStyle, &resp.Diagnostics); ok {
		newState.Git.UsernameStyle = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/submoduleCheckout", oldState.Git.Submodules, plan.Git.Submodules, &resp.Diagnostics); ok {
		newState.Git.Submodules = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/userForTags", oldState.Git.UsernameForTags, plan.Git.UsernameForTags, &resp.Diagnostics); ok {
		newState.Git.UsernameForTags = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/authMethod", oldState.Git.AuthMethod, plan.Git.AuthMethod, &resp.Diagnostics); ok {
		newState.Git.AuthMethod = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/username", oldState.Git.Username, plan.Git.Username, &resp.Diagnostics); ok {
		newState.Git.Username = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/secure:password", oldState.Git.Password, plan.Git.Password, &resp.Diagnostics); ok {
		newState.Git.Password = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/teamcitySshKey", oldState.Git.UploadedKey, plan.Git.UploadedKey, &resp.Diagnostics); ok {
		newState.Git.UploadedKey = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/privateKeyPath", oldState.Git.PrivateKeyPath, plan.Git.PrivateKeyPath, &resp.Diagnostics); ok {
		newState.Git.PrivateKeyPath = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/secure:passphrase", oldState.Git.Passphrase, plan.Git.Passphrase, &resp.Diagnostics); ok {
		newState.Git.Passphrase = result
	} else {
		return
	}

	if result, ok := r.setFieldBool(resourceId, "properties/ignoreKnownHosts", oldState.Git.IgnoreKnownHosts, plan.Git.IgnoreKnownHosts, &resp.Diagnostics); ok {
		newState.Git.IgnoreKnownHosts = result
	} else {
		return
	}

	if result, ok := r.setFieldBool(resourceId, "properties/serverSideAutoCrlf", oldState.Git.ConvertCrlf, plan.Git.ConvertCrlf, &resp.Diagnostics); ok {
		newState.Git.ConvertCrlf = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/agentGitPath", oldState.Git.PathToGit, plan.Git.PathToGit, &resp.Diagnostics); ok {
		newState.Git.PathToGit = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/useAlternates", oldState.Git.CheckoutPolicy, plan.Git.CheckoutPolicy, &resp.Diagnostics); ok {
		newState.Git.CheckoutPolicy = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/agentCleanPolicy", oldState.Git.CleanPolicy, plan.Git.CleanPolicy, &resp.Diagnostics); ok {
		newState.Git.CleanPolicy = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "properties/agentCleanFilesPolicy", oldState.Git.CleanFilesPolicy, plan.Git.CleanFilesPolicy, &resp.Diagnostics); ok {
		newState.Git.CleanFilesPolicy = result
	} else {
		return
	}

	if result, ok := r.setFieldString(resourceId, "id", oldState.Id, plan.Id, &resp.Diagnostics); ok {
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

func (r *vcsRootResource) setFieldString(id, name string, state, plan types.String, diag *diag.Diagnostics) (types.String, bool) {
	if plan.Equal(state) {
		return state, true
	}

	var strVal *string
	if plan.IsNull() {
		strVal = nil
	} else {
		val := plan.ValueString()
		strVal = &val
	}

	result, err := r.client.SetField("vcs-roots", id, name, strVal)
	if err != nil {
		diag.AddError(
			"Error setting VCS root field",
			err.Error(),
		)
		return types.String{}, false
	}

	if *result != "" {
		return types.StringValue(*result), true
	} else {
		return types.StringNull(), true
	}
}

func (r *vcsRootResource) setFieldInt(id, name string, state, plan types.Int64, diag *diag.Diagnostics) (types.Int64, bool) {
	if plan.Equal(state) {
		return state, true
	}

	var strVal *string
	if plan.IsNull() {
		val := ""
		strVal = &val
	} else {
		val := strconv.FormatInt(plan.ValueInt64(), 10)
		strVal = &val
	}

	result, err := r.client.SetField("vcs-roots", id, name, strVal)
	if err != nil {
		diag.AddError(
			"Error setting VCS root field",
			err.Error(),
		)
		return types.Int64{}, false
	}

	intVal, err := strconv.ParseInt(*result, 10, 64)
	if err != nil {
		diag.AddError(
			"Error setting VCS root field",
			err.Error(),
		)
		return types.Int64{}, false
	}
	return types.Int64Value(intVal), true
}

func (r *vcsRootResource) setFieldBool(id, name string, state, plan types.Bool, diag *diag.Diagnostics) (types.Bool, bool) {
	if plan.Equal(state) {
		return state, true
	}

	var strVal *string
	if plan.IsNull() {
		strVal = nil
	} else {
		val := strconv.FormatBool(plan.ValueBool())
		strVal = &val
	}

	result, err := r.client.SetField("vcs-roots", id, name, strVal)
	if err != nil {
		diag.AddError(
			"Error setting VCS root field",
			err.Error(),
		)
		return types.Bool{}, false
	}

	if *result == "" {
		return types.BoolNull(), true
	}

	val, err := strconv.ParseBool(*result)
	if err != nil {
		diag.AddError(
			"Error setting VCS root field",
			err.Error(),
		)
		return types.Bool{}, false
	}
	return types.BoolValue(val), true
}

func (r *vcsRootResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vcsRootResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVcsRoot(state.Id.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Deleting VCS root",
			err.Error(),
		)
		return
	}
}
