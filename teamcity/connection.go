package teamcity

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"
)

var (
	_ resource.Resource              = &connectionResource{}
	_ resource.ResourceWithConfigure = &connectionResource{}
)

func NewConnectionResource() resource.Resource {
	return &connectionResource{}
}

type connectionResource struct {
	client *client.Client
}

type connectionResourceModel struct {
	ProjectId types.String `tfsdk:"project_id"`
	FeatureId types.String `tfsdk:"feature_id"`
	GithubApp GithubApp    `tfsdk:"github_app"`
}

type GithubApp struct {
	DisplayName   types.String `tfsdk:"display_name"`
	OwnerUrl      types.String `tfsdk:"owner_url"`
	AppId         types.String `tfsdk:"app_id"`
	ClientId      types.String `tfsdk:"client_id"`
	ClientSecret  types.String `tfsdk:"client_secret"`
	PrivateKey    types.String `tfsdk:"private_key"`
	WebhookSecret types.String `tfsdk:"webhook_secret"`
}

func (r *connectionResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_connection"
}

func (r *connectionResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "TeamCity allows storing presets of connections to external services. Currently only GitHub App type is supported for adding SSO to the server. More info [here](https://www.jetbrains.com/help/teamcity/configuring-connections.html#GitHub)",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"feature_id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"github_app": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"display_name": schema.StringAttribute{
						Required: true,
					},
					"owner_url": schema.StringAttribute{
						Required: true,
					},
					"app_id": schema.StringAttribute{
						Required: true,
					},
					"client_id": schema.StringAttribute{
						Required: true,
					},
					"client_secret": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
					"private_key": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
					"webhook_secret": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
				},
			},
		},
	}
}

func (r *connectionResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *connectionResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan connectionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	feature := models.ProjectFeatureJson{
		Type: "OAuthProvider",
		Properties: models.Properties{
			Property: []models.Property{
				{
					Name:  "providerType",
					Value: "GitHubApp",
				},
				{
					Name:  "connectionSubtype",
					Value: "gitHubApp",
				},
				{
					Name:  "displayName",
					Value: plan.GithubApp.DisplayName.ValueString(),
				},
				{
					Name:  "gitHubApp.ownerUrl",
					Value: plan.GithubApp.OwnerUrl.ValueString(),
				},
				{
					Name:  "gitHubApp.appId",
					Value: plan.GithubApp.AppId.ValueString(),
				},
				{
					Name:  "gitHubApp.clientId",
					Value: plan.GithubApp.ClientId.ValueString(),
				},
				{
					Name:  "secure:gitHubApp.clientSecret",
					Value: plan.GithubApp.ClientSecret.ValueString(),
				},
				{
					Name:  "secure:gitHubApp.privateKey",
					Value: plan.GithubApp.PrivateKey.ValueString(),
				},
				{
					Name:  "secure:gitHubApp.webhookSecret",
					Value: plan.GithubApp.WebhookSecret.ValueString(),
				},
			},
		},
	}

	result, err := r.client.NewProjectFeature(plan.ProjectId.ValueString(), feature)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding project feature",
			err.Error(),
		)
		return
	}

	newState := r.readState(result, plan)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *connectionResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState connectionResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetProjectFeature(oldState.ProjectId.ValueString(), oldState.FeatureId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error reading connection",
			err.Error(),
		)
		return
	}

	if result == nil {
		resp.State.RemoveResource(ctx)
		return
	}

	newState := r.readState(*result, oldState)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *connectionResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan connectionResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState connectionResourceModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var newState connectionResourceModel
	newState.ProjectId = plan.ProjectId
	newState.FeatureId = plan.FeatureId
	projectId := plan.ProjectId.ValueString()
	featureId := plan.FeatureId.ValueString()

	if result, ok := r.setFieldString(projectId, featureId, "displayName", oldState.GithubApp.DisplayName, plan.GithubApp.DisplayName, &resp.Diagnostics); ok {
		newState.GithubApp.DisplayName = result
	} else {
		return
	}

	if result, ok := r.setFieldString(projectId, featureId, "gitHubApp.ownerUrl", oldState.GithubApp.OwnerUrl, plan.GithubApp.OwnerUrl, &resp.Diagnostics); ok {
		newState.GithubApp.OwnerUrl = result
	} else {
		return
	}

	if result, ok := r.setFieldString(projectId, featureId, "gitHubApp.appId", oldState.GithubApp.AppId, plan.GithubApp.AppId, &resp.Diagnostics); ok {
		newState.GithubApp.AppId = result
	} else {
		return
	}

	if result, ok := r.setFieldString(projectId, featureId, "gitHubApp.clientId", oldState.GithubApp.ClientId, plan.GithubApp.ClientId, &resp.Diagnostics); ok {
		newState.GithubApp.ClientId = result
	} else {
		return
	}

	if result, ok := r.setFieldString(projectId, featureId, "secure:gitHubApp.clientSecret", oldState.GithubApp.ClientSecret, plan.GithubApp.ClientSecret, &resp.Diagnostics); ok {
		newState.GithubApp.ClientSecret = result
	} else {
		return
	}

	if result, ok := r.setFieldString(projectId, featureId, "secure:gitHubApp.privateKey", oldState.GithubApp.PrivateKey, plan.GithubApp.PrivateKey, &resp.Diagnostics); ok {
		newState.GithubApp.PrivateKey = result
	} else {
		return
	}

	if result, ok := r.setFieldString(projectId, featureId, "secure:gitHubApp.webhookSecret", oldState.GithubApp.WebhookSecret, plan.GithubApp.WebhookSecret, &resp.Diagnostics); ok {
		newState.GithubApp.WebhookSecret = result
	} else {
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *connectionResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state connectionResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteProjectFeature(state.ProjectId.ValueString(), state.FeatureId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting connection",
			err.Error(),
		)
		return
	}
}

func (r *connectionResource) readState(result models.ProjectFeatureJson, plan connectionResourceModel) connectionResourceModel {
	props := make(map[string]string)
	for _, p := range result.Properties.Property {
		props[p.Name] = p.Value
	}

	var newState connectionResourceModel
	newState.ProjectId = plan.ProjectId
	newState.FeatureId = types.StringValue(*result.Id)
	newState.GithubApp.DisplayName = types.StringValue(props["displayName"])
	newState.GithubApp.OwnerUrl = types.StringValue(props["gitHubApp.ownerUrl"])
	newState.GithubApp.AppId = types.StringValue(props["gitHubApp.appId"])
	newState.GithubApp.ClientId = types.StringValue(props["gitHubApp.clientId"])
	if _, ok := props["secure:gitHubApp.clientSecret"]; ok {
		newState.GithubApp.ClientSecret = plan.GithubApp.ClientSecret
	}
	if _, ok := props["secure:gitHubApp.privateKey"]; ok {
		newState.GithubApp.PrivateKey = plan.GithubApp.PrivateKey
	}
	if _, ok := props["secure:gitHubApp.webhookSecret"]; ok {
		newState.GithubApp.WebhookSecret = plan.GithubApp.WebhookSecret
	}
	return newState
}

func (r *connectionResource) setFieldString(projectId, featureId, name string, state, plan types.String, diag *diag.Diagnostics) (types.String, bool) {
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

	prop := fmt.Sprintf("projectFeatures/%s/properties/%s", featureId, name)
	result, err := r.client.SetField("projects", projectId, prop, strVal)
	if err != nil {
		diag.AddError(
			"Error setting property",
			err.Error(),
		)
		return types.String{}, false
	}

	if result == "" {
		return types.StringNull(), true
	}

	return types.StringValue(result), true
}
