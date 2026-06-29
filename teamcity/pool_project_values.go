package teamcity

import (
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/models"
)

// getProjectsAttrValue converts TeamCity pool members into Terraform set values.
// Virtual projects are auto-generated and inherited into pools, so they must not
// enter Terraform state or they will cause perpetual drift. Projects without an
// ID are skipped defensively as they cannot be represented in provider state.
func getProjectsAttrValue(data []models.ProjectJson) []attr.Value {
	projects := []attr.Value{}
	for _, p := range data {
		if p.Virtual || p.Id == nil {
			continue
		}
		projects = append(projects, types.StringValue(*p.Id))
	}

	return projects
}
