package teamcity

import (
	"context"
	"fmt"
	"strings"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringdefault"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

var (
	_ resource.Resource                = &paramResource{}
	_ resource.ResourceWithConfigure   = &paramResource{}
	_ resource.ResourceWithImportState = &paramResource{}
)

func NewParamResource() resource.Resource {
	return &paramResource{}
}

type paramResource struct {
	client *client.Client
}

type paramResourceModel struct {
	Id        types.String `tfsdk:"id"`
	ProjectId types.String `tfsdk:"project_id"`
	Name      types.String `tfsdk:"name"`
	Value     types.String `tfsdk:"value"`
	Type      types.String `tfsdk:"type"`
}

func (r *paramResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project_parameter"
}

func (r *paramResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Parameters are name=value pairs that can be referenced throughout TeamCity. Project parameters are available to any Build Configuration inside corresponding project. More info [here](https://www.jetbrains.com/help/teamcity/configuring-build-parameters.html)",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed:    true,
				Description: "Resource identifier in the form 'project_id/name'.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"project_id": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				Required: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"value": schema.StringAttribute{
				Required:  true,
				Sensitive: true,
			},
			"type": schema.StringAttribute{
				Optional: true,
				Computed: true,
				Default:  stringdefault.StaticString(models.ParamTypeText),
				Validators: []validator.String{
					stringvalidator.OneOf([]string{models.ParamTypeText, models.ParamTypePassword}...),
				},
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
					stringplanmodifier.RequiresReplace(),
				},
				Description: "Parameter type. Use 'password' to create a secure (hidden) parameter. Defaults to 'text' if omitted.",
			},
		},
	}
}

func (r *paramResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *paramResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan paramResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := plan.Name.ValueString()
	var err error
	if strings.EqualFold(plan.Type.ValueString(), models.ParamTypePassword) {
		err = r.client.SecureSetParam(plan.ProjectId.ValueString(), name, plan.Value.ValueString())
	} else {
		err = r.client.SetParam(plan.ProjectId.ValueString(), name, plan.Value.ValueString())
	}
	if err != nil {
		resp.Diagnostics.AddError(
			"Error adding project parameter",
			err.Error(),
		)
		return
	}

	var newState paramResourceModel
	newState.Id = types.StringValue(fmt.Sprintf("%s/%s", plan.ProjectId.ValueString(), plan.Name.ValueString()))
	newState.ProjectId = plan.ProjectId
	newState.Name = plan.Name
	newState.Value = plan.Value
	if plan.Type.IsNull() || plan.Type.ValueString() == "" {
		newState.Type = types.StringValue(models.ParamTypeText)
	} else {
		newState.Type = plan.Type
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *paramResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var oldState paramResourceModel
	diags := req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := oldState.Name.ValueString()
	var newState paramResourceModel
	newState.Id = types.StringValue(fmt.Sprintf("%s/%s", oldState.ProjectId.ValueString(), name))
	newState.ProjectId = oldState.ProjectId
	newState.Name = oldState.Name

	isPassword := strings.EqualFold(oldState.Type.ValueString(), models.ParamTypePassword)
	if isPassword {
		// Server does not return secure value; keep it from state to avoid unwanted diffs
		newState.Value = oldState.Value
		newState.Type = types.StringValue(models.ParamTypePassword)
	} else {
		result, err := r.client.GetParam(oldState.ProjectId.ValueString(), name)
		if err != nil {
			resp.Diagnostics.AddError(
				"Error reading group param",
				err.Error(),
			)
			return
		}

		if result == nil {
			resp.State.RemoveResource(ctx)
			return
		}

		newState.Value = types.StringValue(*result)
		if oldState.Type.IsNull() || oldState.Type.ValueString() == "" {
			newState.Type = types.StringValue(models.ParamTypeText)
		} else {
			newState.Type = oldState.Type
		}
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *paramResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan paramResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var oldState paramResourceModel
	diags = req.State.Get(ctx, &oldState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if !plan.Value.Equal(oldState.Value) {
		name := plan.Name.ValueString()
		var err error
		if strings.EqualFold(plan.Type.ValueString(), models.ParamTypePassword) {
			err = r.client.SecureSetParam(plan.ProjectId.ValueString(), name, plan.Value.ValueString())
		} else {
			err = r.client.SetParam(plan.ProjectId.ValueString(), name, plan.Value.ValueString())
		}
		if err != nil {
			resp.Diagnostics.AddError(
				"Error updating project param",
				err.Error(),
			)
			return
		}
	}

	var newState paramResourceModel
	newState.Id = types.StringValue(fmt.Sprintf("%s/%s", plan.ProjectId.ValueString(), plan.Name.ValueString()))
	newState.ProjectId = plan.ProjectId
	newState.Name = plan.Name
	newState.Value = plan.Value
	if plan.Type.IsNull() || plan.Type.ValueString() == "" {
		newState.Type = types.StringValue(models.ParamTypeText)
	} else {
		newState.Type = plan.Type
	}

	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *paramResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state paramResourceModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	name := state.Name.ValueString()
	err := r.client.DeleteParam(state.ProjectId.ValueString(), name)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error deleting project param",
			err.Error(),
		)
		return
	}
}

func (r *paramResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	idParts := strings.Split(req.ID, "/")

	if len(idParts) != 2 || idParts[0] == "" || idParts[1] == "" {
		resp.Diagnostics.AddError(
			"Unexpected Import Identifier",
			fmt.Sprintf("Expected import identifier with format: project_id/name. Got: %q", req.ID),
		)
		return
	}

	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("id"), req.ID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_id"), idParts[0])...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("name"), idParts[1])...)
	// Default type for imported parameters is text; users can change it to password if needed
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("type"), models.ParamTypeText)...)
}
