package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource              = &emailResource{}
	_ resource.ResourceWithConfigure = &emailResource{}
)

func NewEmailResource() resource.Resource {
	return &emailResource{}
}

type emailResource struct {
	client *client.Client
}

type emailResourceModel struct {
	Enabled          types.Bool   `tfsdk:"enabled"`
	Host             types.String `tfsdk:"host"`
	Port             types.Int64  `tfsdk:"port"`
	From             types.String `tfsdk:"from"`
	Login            types.String `tfsdk:"login"`
	Password         types.String `tfsdk:"password"`
	SecureConnection types.String `tfsdk:"secure_connection"`
}

func (r *emailResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *emailResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_email_settings"
}

func (r *emailResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"enabled":  schema.BoolAttribute{Required: true},
			"host":     schema.StringAttribute{Required: true},
			"port":     schema.Int64Attribute{Required: true},
			"from":     schema.StringAttribute{Required: true},
			"login":    schema.StringAttribute{Required: true},
			"password": schema.StringAttribute{Required: true, Sensitive: true},
			"secure_connection": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"NONE", "STARTTLS", "SSL"}...),
				},
			},
		},
	}
}

func (r *emailResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan emailResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting email settings",
			err.Error(),
		)
		return
	}
	newState.Password = plan.Password

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *emailResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState emailResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	result, err := r.client.GetEmailSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read email settings",
			err.Error(),
		)
		return
	}

	newState, err := r.readState(*result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read email settings",
			err.Error(),
		)
		return
	}
	newState.Password = oldState.Password

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *emailResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan emailResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting email settings",
			err.Error(),
		)
		return
	}
	newState.Password = plan.Password

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *emailResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *emailResource) update(plan emailResourceModel) (*emailResourceModel, error) {
	password := plan.Password.ValueString()
	settings := client.EmailSettings{
		Enabled:          plan.Enabled.ValueBool(),
		Host:             plan.Host.ValueString(),
		Port:             int(plan.Port.ValueInt64()),
		From:             plan.From.ValueString(),
		Login:            plan.Login.ValueString(),
		Password:         &password,
		SecureConnection: plan.SecureConnection.ValueString(),
	}

	result, err := r.client.SetEmailSettings(settings)
	if err != nil {
		return nil, err
	}

	return r.readState(*result)
}

func (r *emailResource) readState(result client.EmailSettings) (*emailResourceModel, error) {
	var state emailResourceModel

	state.Enabled = types.BoolValue(result.Enabled)
	state.Host = types.StringValue(result.Host)
	state.Port = types.Int64Value(int64(result.Port))
	state.From = types.StringValue(result.From)
	state.Login = types.StringValue(result.Login)
	state.SecureConnection = types.StringValue(result.SecureConnection)

	return &state, nil
}
