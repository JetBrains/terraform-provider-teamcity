package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/booldefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/int64default"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/objectdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                = &globalResource{}
	_ resource.ResourceWithConfigure   = &globalResource{}
	_ resource.ResourceWithImportState = &globalResource{}
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
	Enabled      types.Bool   `tfsdk:"enabled"`
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
		Description: "[General TeamCity configuration](https://www.jetbrains.com/help/teamcity/teamcity-configuration-and-maintenance.html#TeamCity+Configuration)",
		Attributes: map[string]schema.Attribute{
			"artifact_directories": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("system/artifacts"),
				Description: "New-line delimited. Default: `system/artifacts`",
			},
			"root_url": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Default:     stringdefault.StaticString("http://localhost:8111"),
				Description: "Server URL. Default: `\"http://localhost:8111\"`",
			},
			"max_artifact_size": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(314572800),
				Description: "Maximum build artifact file size. Default: `300 MB`",
			},
			"max_artifact_number": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(1000),
				Description: "Maximum number of artifacts per build. Default: `1000`",
			},
			"default_execution_timeout": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(0),
				Description: "Default build execution timeout. In minutes, default: `0`",
			},
			"default_vcs_check_interval": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "Default VCS changes check interval. In seconds, default: `60`",
			},
			"enforce_default_vcs_check_interval": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Default:     booldefault.StaticBool(false),
				Description: "Enforce this value as a minimum for all VCS Roots. Default: `false`",
			},
			"default_quiet_period": schema.Int64Attribute{
				Optional:    true,
				Computed:    true,
				Default:     int64default.StaticInt64(60),
				Description: "Default VCS trigger quiet period. In seconds, default: `60`",
			},
			"encryption": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Use a custom encryption key. Leave it out to keep the server's current encryption settings, including a key the server generated at first start (TeamCity 2026.2 does). Removing it after it was set switches the server back to TeamCity's default encryption.",
				Attributes: map[string]schema.Attribute{
					"key": schema.StringAttribute{
						Required:    true,
						Sensitive:   true,
						Description: "128 bit key encoded with Base64",
						// An empty key is not sent, so TeamCity would keep the key it has.
						Validators: []validator.String{
							stringvalidator.LengthAtLeast(1),
						},
					},
				},
				Default: objectdefault.StaticValue(types.ObjectNull(
					map[string]attr.Type{"key": types.StringType},
				)),
			},
			"artifacts_domain_isolation": schema.SingleNestedAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Artifacts Domain isolation protection",
				Attributes: map[string]schema.Attribute{
					"enabled": schema.BoolAttribute{
						Optional:    true,
						Computed:    true,
						Default:     booldefault.StaticBool(true),
						Description: "Enabled by default, set false if it needs to be disabled",
					},
					"artifacts_url": schema.StringAttribute{
						Optional:    true,
						Computed:    true,
						Description: "a URL to serve build artifacts from. The URL must be different from the Server URL",
					},
				},
				Default: objectdefault.StaticValue(types.ObjectValueMust(
					map[string]attr.Type{
						"enabled":       types.BoolType,
						"artifacts_url": types.StringType,
					},
					map[string]attr.Value{
						"enabled":       types.BoolValue(true),
						"artifacts_url": types.StringValue(""),
					},
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

	newState, err := r.update(plan, nil)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting global settings",
			err.Error(),
		)
		return
	}
	newState.Encryption = configuredEncryption(newState.Encryption, plan.Encryption)

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
	newState.Encryption = configuredEncryption(newState.Encryption, oldState.Encryption)

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

	var state globalResourceModel
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan, state.Encryption)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting global settings",
			err.Error(),
		)
		return
	}
	newState.Encryption = configuredEncryption(newState.Encryption, plan.Encryption)

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *globalResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *globalResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.Set(ctx, globalResourceModel{})
}

func (r *globalResource) update(plan globalResourceModel, priorEncryption *EncryptionModel) (*globalResourceModel, error) {
	useEncryption, key := r.encryptionToSet(plan.Encryption, priorEncryption)

	enableArtifactsDomainIsolation := true
	var artifactsURL string
	if plan.ArtifactIsolation != nil {
		enableArtifactsDomainIsolation = plan.ArtifactIsolation.Enabled.ValueBool()
		artifactsURL = plan.ArtifactIsolation.ArtifactsUrl.ValueString()
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
		UseEncryption:                  useEncryption,
		EncryptionKey:                  key,
		ArtifactsDomainIsolation:       enableArtifactsDomainIsolation,
		ArtifactsUrl:                   artifactsURL,
	}

	result, err := r.client.SetGlobalSettings(settings)
	if err != nil {
		return nil, err
	}

	return r.readState(*result)
}

// encryptionToSet returns the useEncryption and encryptionKey values to send. A configured block
// sets its key, and removing the block turns the custom key off. With no block in either the
// configuration or the prior state, useEncryption is omitted so TeamCity leaves its current
// encryption settings untouched (both on TeamCity 2026.2, which generates a custom key at first
// start, and on TeamCity <= 2026.1.2, which rejects useEncryption=true without encryptionKey).
func (r *globalResource) encryptionToSet(configured, prior *EncryptionModel) (*bool, string) {
	if configured != nil {
		use := true
		return &use, configured.Key.ValueString()
	}
	if prior != nil {
		use := false
		return &use, ""
	}
	return nil, ""
}

// configuredEncryption returns the encryption block to keep in the state. TeamCity never returns
// the key, so the block holds the configured one while the server uses a custom key. Without a
// configured block the state has none either, whatever the server uses, so that leaving the block
// out plans no change.
func configuredEncryption(server, configured *EncryptionModel) *EncryptionModel {
	if server == nil || configured == nil {
		return nil
	}
	return &EncryptionModel{Key: configured.Key}
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

	if result.UseEncryption != nil && *result.UseEncryption {
		state.Encryption = &EncryptionModel{}
	}

	state.ArtifactIsolation = &ArtifactIsolationModel{
		Enabled:      types.BoolValue(result.ArtifactsDomainIsolation),
		ArtifactsUrl: types.StringValue(result.ArtifactsUrl),
	}

	return &state, nil
}
