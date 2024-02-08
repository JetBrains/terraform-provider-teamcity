package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/objectvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
)

var (
	_ resource.Resource                = &cleanupResource{}
	_ resource.ResourceWithConfigure   = &cleanupResource{}
	_ resource.ResourceWithImportState = &cleanupResource{}
)

func NewCleanupResource() resource.Resource {
	return &cleanupResource{}
}

type cleanupResource struct {
	client *client.Client
}

type cleanupResourceModel struct {
	Enabled     types.Bool          `tfsdk:"enabled"`
	MaxDuration types.Int64         `tfsdk:"max_duration"`
	Daily       *dailyResourceModel `tfsdk:"daily"`
	Cron        *cronResourceModel  `tfsdk:"cron"`
}

type dailyResourceModel struct {
	Hour   types.Int64 `tfsdk:"hour"`
	Minute types.Int64 `tfsdk:"minute"`
}

type cronResourceModel struct {
	Minute  types.String `tfsdk:"minute"`
	Hour    types.String `tfsdk:"hour"`
	Day     types.String `tfsdk:"day"`
	Month   types.String `tfsdk:"month"`
	DayWeek types.String `tfsdk:"day_week"`
}

func (r *cleanupResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cleanup_settings"
}

func (r *cleanupResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "[TeamCity clean-up](https://www.jetbrains.com/help/teamcity/teamcity-data-clean-up.html) functionality allows an automatic deletion of old and no longer necessary build data.",
		Attributes: map[string]schema.Attribute{
			"enabled": schema.BoolAttribute{
				Required: true,
			},
			"max_duration": schema.Int64Attribute{
				Required: true,
			},
			"daily": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"hour": schema.Int64Attribute{
						Required: true,
					},
					"minute": schema.Int64Attribute{
						Required: true,
					},
				},
			},
			"cron": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"minute": schema.StringAttribute{
						Required: true,
					},
					"hour": schema.StringAttribute{
						Required: true,
					},
					"day": schema.StringAttribute{
						Required: true,
					},
					"month": schema.StringAttribute{
						Required: true,
					},
					"day_week": schema.StringAttribute{
						Required: true,
					},
				},

				Validators: []validator.Object{
					objectvalidator.ExactlyOneOf(
						path.MatchRoot("daily"),
						path.MatchRoot("cron"),
					),
				},
			},
		},
	}
}

func (r *cleanupResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *cleanupResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan cleanupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting cleanup",
			"Cannot set cleanup, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	result, err := r.client.GetCleanup()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Cleanup",
			"Could not read cleanup settings: "+err.Error(),
		)
		return
	}

	newState := r.readState(result)
	diags := resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan cleanupResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting cleanup",
			"Cannot set cleanup, unexpected error: "+err.Error(),
		)
		return
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *cleanupResource) ImportState(ctx context.Context, _ resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resp.State.Set(ctx, cleanupResourceModel{})
}

func (r *cleanupResource) update(plan cleanupResourceModel) (cleanupResourceModel, error) {
	settings := client.CleanupSettings{
		Enabled:     plan.Enabled.ValueBool(),
		MaxDuration: int(plan.MaxDuration.ValueInt64()),
	}

	if plan.Daily != nil {
		settings.Daily = &client.CleanupDaily{
			Hour:   int(plan.Daily.Hour.ValueInt64()),
			Minute: int(plan.Daily.Minute.ValueInt64()),
		}
	}
	if plan.Cron != nil {
		settings.Cron = &client.CleanupCron{
			Minute:  plan.Cron.Minute.ValueString(),
			Hour:    plan.Cron.Hour.ValueString(),
			Day:     plan.Cron.Day.ValueString(),
			Month:   plan.Cron.Month.ValueString(),
			DayWeek: plan.Cron.DayWeek.ValueString(),
		}
	}

	result, err := r.client.SetCleanup(settings)
	if err != nil {
		return cleanupResourceModel{}, err
	}

	return r.readState(result), nil
}

func (r *cleanupResource) readState(result client.CleanupSettings) cleanupResourceModel {
	var state cleanupResourceModel

	state.Enabled = types.BoolValue(result.Enabled)
	state.MaxDuration = types.Int64Value(int64(result.MaxDuration))

	if result.Daily != nil {
		state.Daily = &dailyResourceModel{
			Hour:   types.Int64Value(int64(result.Daily.Hour)),
			Minute: types.Int64Value(int64(result.Daily.Minute)),
		}
	}
	if result.Cron != nil {
		state.Cron = &cronResourceModel{
			Minute:  types.StringValue(result.Cron.Minute),
			Hour:    types.StringValue(result.Cron.Hour),
			Day:     types.StringValue(result.Cron.Day),
			Month:   types.StringValue(result.Cron.Month),
			DayWeek: types.StringValue(result.Cron.DayWeek),
		}
	}

	return state
}
