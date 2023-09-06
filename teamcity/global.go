package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource              = &globalResource{}
	_ resource.ResourceWithConfigure = &globalResource{}
)

func NewGlobalResource() resource.Resource {
	return &globalResource{}
}

type globalResource struct {
	client *client.Client
}

type globalResourceModel struct {
	ArtifactDirectories            types.String            `tfsdk:"artifact_directories"`
	RootUrl                        types.String            `tfsdk:"root_url"`
	MaxArtifactSize                types.Int64             `tfsdk:"max_artifact_size"`
	MaxArtifactNumber              types.Int64             `tfsdk:"max_artifact_number"`
	DefaultExecutionTimeout        types.Int64             `tfsdk:"default_execution_timeout"`
	DefaultVCSCheckInterval        types.Int64             `tfsdk:"default_vcs_check_interval"`
	EnforceDefaultVCSCheckInterval types.Bool              `tfsdk:"enforce_default_vcs_check_interval"`
	DefaultQuietPeriod             types.Int64             `tfsdk:"default_quiet_period"`
	Encryption                     *EncryptionModel        `tfsdk:"encryption"`
	ArtifactIsolation              *ArtifactIsolationModel `tfsdk:"artifacts_domain_isolation"`
}

type EncryptionModel struct {
	Key types.String `tfsdk:"key"`
}

type ArtifactIsolationModel struct {
	ArtifactsUrl types.String `tfsdk:"artifacts_url"`
}

func (r *globalResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *globalResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_global_settings"
}

func (r *globalResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"artifact_directories": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("system/artifacts"),
			},
			"root_url": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString("http://localhost:8111"),
			},
			"max_artifact_size": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(314572800),
			},
			"max_artifact_number": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(1000),
			},
			"default_execution_timeout": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(0),
			},
			"default_vcs_check_interval": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
			},
			"enforce_default_vcs_check_interval": schema.BoolAttribute{
				Optional: true,
				Computed: true,
				Default:  booldefault.StaticBool(false),
			},
			"default_quiet_period": schema.Int64Attribute{
				Optional: true,
				Computed: true,
				Default:  int64default.StaticInt64(60),
			},
			"encryption": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"key": schema.StringAttribute{
						Required:  true,
						Sensitive: true,
					},
				},
				Default: objectdefault.StaticValue(types.ObjectNull(
					map[string]attr.Type{"key": types.StringType},
				)),
			},
			"artifacts_domain_isolation": schema.SingleNestedAttribute{
				Optional: true,
				Computed: true,
				Attributes: map[string]schema.Attribute{
					"artifacts_url": schema.StringAttribute{
						Required: true,
					},
				},
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{"artifacts_url": types.StringType},
					map[string]attr.Value{"artifacts_url": types.StringValue("")},
				)),
			},
		},
	}
}

func (r *globalResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan globalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting global settings",
			err.Error(),
		)
		return
	}
	if newState.Encryption != nil {
		newState.Encryption.Key = plan.Encryption.Key
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *globalResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState globalResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetGlobalSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read global settings",
			err.Error(),
		)
		return
	}

	newState, err := r.readState(*result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read global settings",
			err.Error(),
		)
		return
	}
	if newState.Encryption != nil {
		newState.Encryption.Key = oldState.Encryption.Key
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *globalResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan globalResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting global settings",
			err.Error(),
		)
		return
	}
	if newState.Encryption != nil {
		newState.Encryption.Key = plan.Encryption.Key
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *globalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *globalResource) update(plan globalResourceModel) (*globalResourceModel, error) {
	var key string
	if plan.Encryption != nil {
		v := plan.Encryption.Key.ValueString()
		key = v
	}

	var url string
	if plan.ArtifactIsolation != nil {
		v := plan.ArtifactIsolation.ArtifactsUrl.ValueString()
		url = v
	}

	settings := client.GlobalSettings{
		ArtifactDirectories:            plan.ArtifactDirectories.ValueString(),
		RootUrl:                        plan.RootUrl.ValueString(),
		MaxArtifactSize:                plan.MaxArtifactSize.ValueInt64(),
		MaxArtifactNumber:              plan.MaxArtifactNumber.ValueInt64(),
		DefaultExecutionTimeout:        plan.DefaultExecutionTimeout.ValueInt64(),
		DefaultVCSCheckInterval:        plan.DefaultVCSCheckInterval.ValueInt64(),
		EnforceDefaultVCSCheckInterval: plan.EnforceDefaultVCSCheckInterval.ValueBool(),
		DefaultQuietPeriod:             plan.DefaultQuietPeriod.ValueInt64(),
		UseEncryption:                  plan.Encryption != nil,
		EncryptionKey:                  key,
		ArtifactsDomainIsolation:       plan.ArtifactIsolation != nil,
		ArtifactsUrl:                   url,
	}

	result, err := r.client.SetGlobalSettings(settings)
	if err != nil {
		return nil, err
	}

	return r.readState(*result)
}

func (r *globalResource) readState(result client.GlobalSettings) (*globalResourceModel, error) {
	var state globalResourceModel

	state.ArtifactDirectories = types.StringValue(result.ArtifactDirectories)
	state.RootUrl = types.StringValue(result.RootUrl)
	state.MaxArtifactSize = types.Int64Value(result.MaxArtifactSize)
	state.MaxArtifactNumber = types.Int64Value(result.MaxArtifactNumber)
	state.DefaultExecutionTimeout = types.Int64Value(result.DefaultExecutionTimeout)
	state.DefaultVCSCheckInterval = types.Int64Value(result.DefaultVCSCheckInterval)
	state.EnforceDefaultVCSCheckInterval = types.BoolValue(result.EnforceDefaultVCSCheckInterval)
	state.DefaultQuietPeriod = types.Int64Value(result.DefaultQuietPeriod)

	if result.UseEncryption {
		state.Encryption = &EncryptionModel{}
	}

	if result.ArtifactsDomainIsolation {
		state.ArtifactIsolation = &ArtifactIsolationModel{
			ArtifactsUrl: types.StringValue(result.ArtifactsUrl),
		}
	}

	return &state, nil
}
