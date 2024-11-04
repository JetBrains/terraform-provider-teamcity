package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type SshKeyDataModel struct {
	Name      types.String `tfsdk:"name"`
	ProjectId types.String `tfsdk:"project_id"`
}
