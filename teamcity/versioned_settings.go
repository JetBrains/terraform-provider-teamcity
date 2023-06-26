package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"strconv"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource              = &versionedSettingsResource{}
	_ resource.ResourceWithConfigure = &versionedSettingsResource{}
)

func NewVersionedSettingsResource() resource.Resource {
	return &versionedSettingsResource{}
}

type versionedSettingsResource struct {
	client *client.Client
}

type versionedSettingsModel struct {
	ProjectId      types.String `tfsdk:"project_id"`
	VcsRoot        types.String `tfsdk:"vcsroot_id"`
	AllowUIEditing types.Bool   `tfsdk:"allow_ui_editing"`
	Settings       types.String `tfsdk:"settings"`
	ShowChanges    types.Bool   `tfsdk:"show_changes"`
}

func (r *versionedSettingsResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_versioned_settings"
}

func (r *versionedSettingsResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"vcsroot_id":       schema.StringAttribute{Required: true},
			"allow_ui_editing": schema.BoolAttribute{Required: true},
			"settings": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"alwaysUseCurrent", "useCurrentByDefault", "useFromVCS"}...),
				},
			},
			"show_changes": schema.BoolAttribute{Required: true},
		},
	}
}

func (r *versionedSettingsResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *versionedSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan versionedSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	root := plan.VcsRoot.ValueString()
	format := "kotlin"
	editing := plan.AllowUIEditing.ValueBool()
	secureValuesOutsideVcs := true
	buildSettings := plan.Settings.ValueString()
	showChanges := plan.ShowChanges.ValueBool()
	decision := "importFromVCS"
	settings := client.VersionedSettings{
		SynchronizationMode:         "enabled",
		VcsRootId:                   &root,
		Format:                      &format,
		AllowUIEditing:              &editing,
		StoreSecureValuesOutsideVcs: &secureValuesOutsideVcs,
		BuildSettingsMode:           &buildSettings,
		ShowSettingsChanges:         &showChanges,
		ImportDecision:              &decision,
	}

	projectId := plan.ProjectId.ValueString()
	result, err := r.client.SetVersionedSettings(projectId, settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting versioned settings",
			err.Error(),
		)
		return
	}

	newState, err := r.readState(*result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading versioned settings",
			err.Error(),
		)
		return
	}
	newState.ProjectId = plan.ProjectId

	diags = resp.State.Set(ctx, *newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *versionedSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState versionedSettingsModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetVersionedSettings(oldState.ProjectId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading versioned settings",
			err.Error(),
		)
		return
	}

	if *actual.Format != "kotlin" {
		resp.State.RemoveResource(ctx)
		return
	}

	newState, err := r.readState(*actual)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading versioned settings",
			err.Error(),
		)
		return
	}
	newState.ProjectId = oldState.ProjectId

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *versionedSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan versionedSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState versionedSettingsModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState versionedSettingsModel
	projectId := plan.ProjectId.ValueString()
	newState.ProjectId = plan.ProjectId

	if result, ok := r.setPropertyString(projectId, "vcsRootId", oldState.VcsRoot, plan.VcsRoot, &resp.Diagnostics); ok {
		newState.VcsRoot = result
	} else {
		return
	}

	if result, ok := r.setPropertyBool(projectId, "allowUIEditing", oldState.AllowUIEditing, plan.AllowUIEditing, &resp.Diagnostics); ok {
		newState.AllowUIEditing = result
	} else {
		return
	}

	if result, ok := r.setPropertyString(projectId, "buildSettings", oldState.Settings, plan.Settings, &resp.Diagnostics); ok {
		newState.Settings = result
	} else {
		return
	}

	if result, ok := r.setPropertyBool(projectId, "showChanges", oldState.ShowChanges, plan.ShowChanges, &resp.Diagnostics); ok {
		newState.ShowChanges = result
	} else {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *versionedSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state versionedSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := client.VersionedSettings{
		SynchronizationMode: "disabled",
	}

	projectId := state.ProjectId.ValueString()
	_, err := r.client.SetVersionedSettings(projectId, settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error disabling versioned settings",
			err.Error(),
		)
		return
	}
}

func (r *versionedSettingsResource) readState(result client.VersionedSettings) (*versionedSettingsModel, error) {
	settings := versionedSettingsModel{
		VcsRoot:        types.StringValue(*result.VcsRootId),
		AllowUIEditing: types.BoolValue(*result.AllowUIEditing),
		Settings:       types.StringValue(*result.BuildSettingsMode),
		ShowChanges:    types.BoolValue(*result.ShowSettingsChanges),
	}

	return &settings, nil
}

func (r *versionedSettingsResource) setPropertyString(projectId, name string, state, plan types.String, diag *diag.Diagnostics) (types.String, bool) {
	if plan.Equal(state) {
		return state, true
	}

	val := plan.ValueString()

	result, err := r.client.SetField("projects", projectId, "versionedSettings/config/parameters/"+name, &val)
	if err != nil {
		diag.AddError(
			"Error setting project feature property",
			err.Error(),
		)
		return types.String{}, false
	}

	return types.StringValue(result), true
}

func (r *versionedSettingsResource) setPropertyBool(projectId, name string, state, plan types.Bool, diag *diag.Diagnostics) (types.Bool, bool) {
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

	result, err := r.client.SetField("projects", projectId, "versionedSettings/config/parameters/"+name, strVal)
	if err != nil {
		diag.AddError(
			"Error setting project feature property",
			err.Error(),
		)
		return types.Bool{}, false
	}

	if result == "" {
		return types.BoolNull(), true
	}

	val, err := strconv.ParseBool(result)
	if err != nil {
		diag.AddError(
			"Error setting field",
			err.Error(),
		)
		return types.Bool{}, false
	}
	return types.BoolValue(val), true
}
