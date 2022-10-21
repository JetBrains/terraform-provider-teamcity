package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/schemavalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"terraform-provider-teamcity/client"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource              = &cleanupResource{}
	_ resource.ResourceWithConfigure = &cleanupResource{}
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
	resp.TypeName = req.ProviderTypeName + "_cleanup"
}

func (r *cleanupResource) GetSchema(_ context.Context) (tfsdk.Schema, diag.Diagnostics) {
	return tfsdk.Schema{
		Attributes: map[string]tfsdk.Attribute{
			"enabled": {
				Type:     types.BoolType,
				Required: true,
			},
			"max_duration": {
				Type:     types.Int64Type,
				Required: true,
			},
			"daily": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"hour": {
						Type:     types.Int64Type,
						Required: true,
					},
					"minute": {
						Type:     types.Int64Type,
						Required: true,
					},
				}),
			},
			"cron": {
				Optional: true,
				Attributes: tfsdk.SingleNestedAttributes(map[string]tfsdk.Attribute{
					"minute": {
						Type:     types.StringType,
						Required: true,
					},
					"hour": {
						Type:     types.StringType,
						Required: true,
					},
					"day": {
						Type:     types.StringType,
						Required: true,
					},
					"month": {
						Type:     types.StringType,
						Required: true,
					},
					"day_week": {
						Type:     types.StringType,
						Required: true,
					},
				}),

				Validators: []tfsdk.AttributeValidator{
					schemavalidator.ExactlyOneOf(
						path.MatchRoot("daily"),
						path.MatchRoot("cron"),
					),
				},
			},
		},
	}, nil
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

	settings := client.CleanupSettings{
		Enabled:     plan.Enabled.Value,
		MaxDuration: int(plan.MaxDuration.Value),
	}

	if plan.Daily != nil {
		settings.Daily = &client.CleanupDaily{
			Hour:   int(plan.Daily.Hour.Value),
			Minute: int(plan.Daily.Minute.Value),
		}
	}
	if plan.Cron != nil {
		settings.Cron = &client.CleanupCron{
			Minute:  plan.Cron.Minute.Value,
			Hour:    plan.Cron.Hour.Value,
			Day:     plan.Cron.Day.Value,
			Month:   plan.Cron.Month.Value,
			DayWeek: plan.Cron.DayWeek.Value,
		}
	}

	result, err := r.client.SetCleanup(settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting cleanup",
			"Cannot set cleanup, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Enabled = types.Bool{Value: result.Enabled}
	plan.MaxDuration = types.Int64{Value: int64(result.MaxDuration)}
	if result.Daily != nil {
		plan.Daily = &dailyResourceModel{
			Hour:   types.Int64{Value: int64(result.Daily.Hour)},
			Minute: types.Int64{Value: int64(result.Daily.Minute)},
		}
	}
	if result.Cron != nil {
		plan.Cron = &cronResourceModel{
			Minute:  types.String{Value: result.Cron.Minute},
			Hour:    types.String{Value: result.Cron.Hour},
			Day:     types.String{Value: result.Cron.Day},
			Month:   types.String{Value: result.Cron.Month},
			DayWeek: types.String{Value: result.Cron.DayWeek},
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state cleanupResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	actual, err := r.client.GetCleanup()
	if err != nil {
		resp.Diagnostics.AddError(
			"Error Reading Cleanup",
			"Could not read cleanup settings: "+err.Error(),
		)
		return
	}

	state.Enabled = types.Bool{Value: actual.Enabled}
	state.MaxDuration = types.Int64{Value: int64(actual.MaxDuration)}
	if actual.Daily != nil {
		state.Daily = &dailyResourceModel{
			Hour:   types.Int64{Value: int64(actual.Daily.Hour)},
			Minute: types.Int64{Value: int64(actual.Daily.Minute)},
		}
		state.Cron = nil
	}

	if actual.Cron != nil {
		state.Cron = &cronResourceModel{
			Minute:  types.String{Value: actual.Cron.Minute},
			Hour:    types.String{Value: actual.Cron.Hour},
			Day:     types.String{Value: actual.Cron.Day},
			Month:   types.String{Value: actual.Cron.Month},
			DayWeek: types.String{Value: actual.Cron.DayWeek},
		}
		state.Daily = nil
	}

	diags = resp.State.Set(ctx, &state)
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

	settings := client.CleanupSettings{
		Enabled:     plan.Enabled.Value,
		MaxDuration: int(plan.MaxDuration.Value),
	}

	if plan.Daily != nil {
		settings.Daily = &client.CleanupDaily{
			Hour:   int(plan.Daily.Hour.Value),
			Minute: int(plan.Daily.Minute.Value),
		}
	}
	if plan.Cron != nil {
		settings.Cron = &client.CleanupCron{
			Minute:  plan.Cron.Minute.Value,
			Hour:    plan.Cron.Hour.Value,
			Day:     plan.Cron.Day.Value,
			Month:   plan.Cron.Month.Value,
			DayWeek: plan.Cron.DayWeek.Value,
		}
	}

	result, err := r.client.SetCleanup(settings)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting cleanup",
			"Cannot set cleanup, unexpected error: "+err.Error(),
		)
		return
	}

	plan.Enabled = types.Bool{Value: result.Enabled}
	plan.MaxDuration = types.Int64{Value: int64(result.MaxDuration)}
	if result.Daily != nil {
		plan.Daily = &dailyResourceModel{
			Hour:   types.Int64{Value: int64(result.Daily.Hour)},
			Minute: types.Int64{Value: int64(result.Daily.Minute)},
		}
	}
	if result.Cron != nil {
		plan.Cron = &cronResourceModel{
			Minute:  types.String{Value: result.Cron.Minute},
			Hour:    types.String{Value: result.Cron.Hour},
			Day:     types.String{Value: result.Cron.Day},
			Month:   types.String{Value: result.Cron.Month},
			DayWeek: types.String{Value: result.Cron.DayWeek},
		}
	}

	diags = resp.State.Set(ctx, plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *cleanupResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
}
