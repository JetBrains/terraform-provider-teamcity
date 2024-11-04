package teamcity

import (
	"context"
	"fmt"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/client"
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/path"
)

var (
	_ datasource.DataSource              = &sshKeyDataSource{}
	_ datasource.DataSourceWithConfigure = &sshKeyDataSource{}
)

func NewSshKeyDataSource() datasource.DataSource {
	return &sshKeyDataSource{}
}

type sshKeyDataSource struct {
	client *client.Client
}

func (d *sshKeyDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_ssh_key"
}

func (d *sshKeyDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "SSH Key from specific Project. More info [here](https://www.jetbrains.com/help/teamcity/ssh-keys-management.html)",
		Attributes: map[string]schema.Attribute{
			"project_id": schema.StringAttribute{
				Required: true,
			},
			"name": schema.StringAttribute{
				Required: true,
			},
		},
	}
}

func (d *sshKeyDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var model models.SshKeyDataModel
	diags := req.Config.Get(ctx, &model)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	sshKeys, err := d.client.GetSshKeys(model.ProjectId.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			fmt.Sprintf("Unable to read SSH keys in project %s", model.ProjectId.ValueString()),
			err.Error(),
		)
		return
	}

	var requestedSshKey = ""
	for _, value := range sshKeys {
		if value == model.Name.ValueString() {
			requestedSshKey = value
		}
	}
	if requestedSshKey == "" {
		resp.Diagnostics.AddAttributeError(
			path.Root("name"),
			"SSH Key not found",
			"The Datasource cannot get SSH Key since there is no SSH Key with the provided name.",
		)
		return
	}

	var state = models.SshKeyDataModel{
		Name:      types.StringValue(requestedSshKey),
		ProjectId: model.ProjectId,
	}
	diags = resp.State.Set(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}
}

func (d *sshKeyDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}

	httpClient, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("Expected *client.Client, got: %T.", req.ProviderData),
		)
		return
	}

	d.client = httpClient
}
