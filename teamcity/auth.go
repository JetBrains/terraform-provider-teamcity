package teamcity

import (
	"context"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/path"
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
	_ resource.Resource                   = &authResource{}
	_ resource.ResourceWithConfigure      = &authResource{}
	_ resource.ResourceWithValidateConfig = &authResource{}
)

func NewAuthResource() resource.Resource {
	return &authResource{}
}

type authResource struct {
	client *client.Client
}

type authResourceModel struct {
	ID                 types.String     `tfsdk:"id"`
	AllowGuest         types.Bool       `tfsdk:"allow_guest"`
	GuestUsername      types.String     `tfsdk:"guest_username"`
	WelcomeText        types.String     `tfsdk:"welcome_text"`
	CollapseLoginForm  types.Bool       `tfsdk:"collapse_login_form"`
	TwoFactorMode      types.String     `tfsdk:"two_factor_mode"`
	ProjectPermissions types.Bool       `tfsdk:"project_permissions"`
	EmailVerification  types.Bool       `tfsdk:"email_verification"`
	Modules            authModulesModel `tfsdk:"modules"`
}

type authModulesModel struct {
	Token            *authModuleTokenModel   `tfsdk:"token"`
	BuiltIn          *authModuleBuiltInModel `tfsdk:"built_in"`
	Google           *authModuleGoogleModel  `tfsdk:"google"`
	GithubCom        *authModuleGithubModel  `tfsdk:"github"`
	GithubEnterprise *authModuleGithubModel  `tfsdk:"github_enterprise"`
	Space            *authModuleSpaceModel   `tfsdk:"jetbrains_space"`
}

func (r *authResource) Configure(_ context.Context, req resource.ConfigureRequest, _ *resource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	r.client = req.ProviderData.(*client.Client)
}

func (r *authResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_auth"
}

func (r *authResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				Computed: true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.UseStateForUnknown(),
				},
			},
			"allow_guest": schema.BoolAttribute{
				Required: true,
			},
			"guest_username": schema.StringAttribute{
				Required: true,
			},
			"welcome_text": schema.StringAttribute{
				Required: true,
			},
			"collapse_login_form": schema.BoolAttribute{
				Required: true,
			},
			"two_factor_mode": schema.StringAttribute{
				Required: true,
				Validators: []validator.String{
					stringvalidator.OneOf([]string{"DISABLED", "OPTIONAL", "MANDATORY"}...),
				},
			},
			"project_permissions": schema.BoolAttribute{
				Required: true,
			},
			"email_verification": schema.BoolAttribute{
				Required: true,
			},
			"modules": schema.SingleNestedAttribute{
				Required: true,
				Attributes: map[string]schema.Attribute{
					"token": schema.SingleNestedAttribute{
						Required: true,
					},
					"built_in": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"registration": schema.BoolAttribute{
								Required: true,
							},
							"change_passwords": schema.BoolAttribute{
								Required: true,
							},
							"reset_passwords": schema.BoolAttribute{
								Optional: true,
							},
						},
					},
					"google": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"create_new_users": schema.BoolAttribute{
								Required: true,
							},
							"all_domains": schema.BoolAttribute{
								Required: true,
							},
							"domains": schema.StringAttribute{
								Optional: true,
							},
						},
					},
					"github": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"create_new_users": schema.BoolAttribute{
								Required: true,
							},
							"organizations": schema.StringAttribute{
								Required: true,
							},
						},
					},
					"github_enterprise": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"create_new_users": schema.BoolAttribute{
								Required: true,
							},
							"organizations": schema.StringAttribute{
								Required: true,
							},
						},
					},
					"jetbrains_space": schema.SingleNestedAttribute{
						Optional: true,
						Attributes: map[string]schema.Attribute{
							"create_new_users": schema.BoolAttribute{
								Required: true,
							},
						},
					},
				},
			},
		},
	}
}

func (r *authResource) ValidateConfig(ctx context.Context, req resource.ValidateConfigRequest, resp *resource.ValidateConfigResponse) {
	var config authResourceModel
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Modules.BuiltIn != nil {
		if config.Modules.BuiltIn.ChangePasswords.ValueBool() == true {
			if config.Modules.BuiltIn.ResetPasswords.IsNull() {
				resp.Diagnostics.AddAttributeError(
					path.Root("modules").AtName("built_in").AtName("reset_passwords"),
					"'reset_passwords' must be specified if 'change_passwords' is enabled",
					"",
				)
			}
		} else {
			if config.Modules.BuiltIn.ResetPasswords.IsNull() != true {
				resp.Diagnostics.AddAttributeError(
					path.Root("modules").AtName("built_in").AtName("reset_passwords"),
					"'reset_passwords' cannot be specified if 'change_passwords' is disabled",
					"",
				)
			}
		}
	}

	if config.Modules.Google != nil {
		if config.Modules.Google.AllDomains.ValueBool() == true {
			if config.Modules.Google.Domains.IsNull() != true {
				resp.Diagnostics.AddAttributeError(
					path.Root("modules").AtName("google").AtName("domains"),
					"'domains' cannot be specified if 'all_domains' is enabled",
					"",
				)
			}
		} else {
			if config.Modules.Google.Domains.IsNull() == true {
				resp.Diagnostics.AddAttributeError(
					path.Root("modules").AtName("google").AtName("domains"),
					"'domains' must be specified if 'all_domains' is disabled",
					"",
				)
			}
		}
	}

}

func (r *authResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan authResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting authentication settings",
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

func (r *authResource) Read(ctx context.Context, _ resource.ReadRequest, resp *resource.ReadResponse) {
	result, err := r.client.GetAuthSettings()
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read authentication settings",
			err.Error(),
		)
		return
	}

	newState, err := r.readState(result)
	if err != nil {
		resp.Diagnostics.AddError(
			"Unable to read authentication settings",
			err.Error(),
		)
		return
	}

	diags := resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *authResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan authResourceModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	newState, err := r.update(plan)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error setting authentication settings",
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

func (r *authResource) Delete(_ context.Context, _ resource.DeleteRequest, _ *resource.DeleteResponse) {
}

func (r *authResource) update(plan authResourceModel) (authResourceModel, error) {
	settings := client.AuthSettings{
		AllowGuest:         plan.AllowGuest.ValueBool(),
		GuestUsername:      plan.GuestUsername.ValueString(),
		WelcomeText:        plan.WelcomeText.ValueString(),
		CollapseLoginForm:  plan.CollapseLoginForm.ValueBool(),
		TwoFactorMode:      plan.TwoFactorMode.ValueString(),
		ProjectPermissions: plan.ProjectPermissions.ValueBool(),
		EmailVerification:  plan.EmailVerification.ValueBool(),
	}

	if plan.Modules.Token != nil {
		settings.Modules.Module = append(settings.Modules.Module, client.Module{Name: "Token-Auth"})
	}
	if plan.Modules.BuiltIn != nil {
		settings.Modules.Module = append(settings.Modules.Module, client.Module{
			Name: "Default",
			Properties: &client.Properties{
				Property: plan.Modules.BuiltIn.getProperties(),
			},
		})
	}
	if plan.Modules.Google != nil {
		settings.Modules.Module = append(settings.Modules.Module, client.Module{
			Name: "Google-oauth",
			Properties: &client.Properties{
				Property: plan.Modules.Google.getProperties(),
			},
		})
	}
	if plan.Modules.GithubCom != nil {
		settings.Modules.Module = append(settings.Modules.Module, client.Module{
			Name: "GitHub-oauth",
			Properties: &client.Properties{
				Property: plan.Modules.GithubCom.getProperties(),
			},
		})
	}
	if plan.Modules.GithubEnterprise != nil {
		settings.Modules.Module = append(settings.Modules.Module, client.Module{
			Name: "GHE-oauth",
			Properties: &client.Properties{
				Property: plan.Modules.GithubEnterprise.getProperties(),
			},
		})
	}
	if plan.Modules.Space != nil {
		settings.Modules.Module = append(settings.Modules.Module, client.Module{
			Name: "JetbrainsSpace-oauth",
			Properties: &client.Properties{
				Property: plan.Modules.Space.getProperties(),
			},
		})
	}

	result, err := r.client.SetAuthSettings(settings)
	if err != nil {
		return authResourceModel{}, err
	}

	return r.readState(result)
}

func (r *authResource) readState(result client.AuthSettings) (authResourceModel, error) {
	var state authResourceModel

	state.ID = types.StringValue("auth")
	state.AllowGuest = types.BoolValue(result.AllowGuest)
	state.GuestUsername = types.StringValue(result.GuestUsername)
	state.WelcomeText = types.StringValue(result.WelcomeText)
	state.CollapseLoginForm = types.BoolValue(result.CollapseLoginForm)
	state.TwoFactorMode = types.StringValue(result.TwoFactorMode)
	state.ProjectPermissions = types.BoolValue(result.ProjectPermissions)
	state.EmailVerification = types.BoolValue(result.EmailVerification)

	for _, module := range result.Modules.Module {
		props := make(map[string]string)
		for _, p := range module.Properties.Property {
			props[p.Name] = p.Value
		}

		if module.Name == "Token-Auth" {
			state.Modules.Token = &authModuleTokenModel{}
			continue
		}

		if module.Name == "Default" {
			state.Modules.BuiltIn = &authModuleBuiltInModel{}
			err := state.Modules.BuiltIn.setFields(props)
			if err != nil {
				return authResourceModel{}, err
			}
			continue
		}

		if module.Name == "Google-oauth" {
			state.Modules.Google = &authModuleGoogleModel{}
			err := state.Modules.Google.setFields(props)
			if err != nil {
				return authResourceModel{}, err
			}
			continue
		}

		if module.Name == "GitHub-oauth" {
			state.Modules.GithubCom = &authModuleGithubModel{}
			err := state.Modules.GithubCom.setFields(props)
			if err != nil {
				return authResourceModel{}, err
			}
			continue
		}

		if module.Name == "GHE-oauth" {
			state.Modules.GithubEnterprise = &authModuleGithubModel{}
			err := state.Modules.GithubEnterprise.setFields(props)
			if err != nil {
				return authResourceModel{}, err
			}
			continue
		}

		if module.Name == "JetbrainsSpace-oauth" {
			state.Modules.Space = &authModuleSpaceModel{}
			err := state.Modules.Space.setFields(props)
			if err != nil {
				return authResourceModel{}, err
			}
			continue
		}
	}

	return state, nil
}

type authModuleTokenModel struct {
}

type authModuleBuiltInModel struct {
	Registration    types.Bool `tfsdk:"registration"`
	ChangePasswords types.Bool `tfsdk:"change_passwords"`
	ResetPasswords  types.Bool `tfsdk:"reset_passwords"`
}

type authModuleGoogleModel struct {
	CreateNewUsers types.Bool   `tfsdk:"create_new_users"`
	AllDomains     types.Bool   `tfsdk:"all_domains"`
	Domains        types.String `tfsdk:"domains"`
}

type authModuleGithubModel struct {
	CreateNewUsers types.Bool   `tfsdk:"create_new_users"`
	Organizations  types.String `tfsdk:"organizations"`
}

type authModuleSpaceModel struct {
	CreateNewUsers types.Bool `tfsdk:"create_new_users"`
}

func (m *authModuleBuiltInModel) getProperties() []client.Property {
	props := []client.Property{
		{Name: "freeRegistrationAllowed", Value: strconv.FormatBool(m.Registration.ValueBool())},
		{Name: "usersCanChangeOwnPasswords", Value: strconv.FormatBool(m.ChangePasswords.ValueBool())},
	}

	if m.ChangePasswords.ValueBool() == true {
		props = append(props, client.Property{
			Name:  "usersCanResetOwnPasswords",
			Value: strconv.FormatBool(m.ResetPasswords.ValueBool()),
		})
	}

	return props
}

func (m *authModuleBuiltInModel) setFields(props map[string]string) error {
	registration, err := strconv.ParseBool(props["freeRegistrationAllowed"])
	if err != nil {
		return err
	}
	change, err := strconv.ParseBool(props["usersCanChangeOwnPasswords"])
	if err != nil {
		return err
	}

	m.Registration = types.BoolValue(registration)
	m.ChangePasswords = types.BoolValue(change)

	if change == true {
		reset, err := strconv.ParseBool(props["usersCanResetOwnPasswords"])
		if err != nil {
			return err
		}
		m.ResetPasswords = types.BoolValue(reset)
	}
	return nil
}

func (m *authModuleGoogleModel) getProperties() []client.Property {
	props := []client.Property{
		{Name: "allowCreatingNewUsersByLogin", Value: strconv.FormatBool(m.CreateNewUsers.ValueBool())},
		{Name: "allowAllUsersToLogin", Value: strconv.FormatBool(m.AllDomains.ValueBool())},
	}

	if m.AllDomains.ValueBool() == false {
		props = append(props, client.Property{
			Name:  "domains",
			Value: m.Domains.ValueString(),
		})
	}

	return props
}

func (m *authModuleGoogleModel) setFields(props map[string]string) error {
	creating, err := strconv.ParseBool(props["allowCreatingNewUsersByLogin"])
	if err != nil {
		return err
	}
	all, err := strconv.ParseBool(props["allowAllUsersToLogin"])
	if err != nil {
		return err
	}

	m.CreateNewUsers = types.BoolValue(creating)
	m.AllDomains = types.BoolValue(all)

	if all == false {
		m.Domains = types.StringValue(props["domains"])
	}

	return nil
}

func (m *authModuleGithubModel) getProperties() []client.Property {
	return []client.Property{
		{Name: "allowCreatingNewUsersByLogin", Value: strconv.FormatBool(m.CreateNewUsers.ValueBool())},
		{Name: "organization", Value: m.Organizations.ValueString()},
	}
}

func (m *authModuleGithubModel) setFields(props map[string]string) error {
	creating, err := strconv.ParseBool(props["allowCreatingNewUsersByLogin"])
	if err != nil {
		return err
	}

	m.CreateNewUsers = types.BoolValue(creating)
	m.Organizations = types.StringValue(props["organization"])
	return nil
}

func (m *authModuleSpaceModel) getProperties() []client.Property {
	return []client.Property{
		{Name: "allowCreatingNewUsersByLogin", Value: strconv.FormatBool(m.CreateNewUsers.ValueBool())},
	}
}

func (m *authModuleSpaceModel) setFields(props map[string]string) error {
	creating, err := strconv.ParseBool(props["allowCreatingNewUsersByLogin"])
	if err != nil {
		return err
	}

	m.CreateNewUsers = types.BoolValue(creating)
	return nil
}
