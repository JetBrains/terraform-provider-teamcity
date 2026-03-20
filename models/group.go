package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type GroupJson struct {
	Key         string               `json:"key,omitempty"`
	Name        string               `json:"name"`
	Description string               `json:"description,omitempty"`
	Roles       *RoleAssignmentsJson `json:"roles,omitempty"`
	Parents     *ParentGroupsJson    `json:"parent-groups,omitempty"`
}

type RoleAssignmentsJson struct {
	RoleAssignment []RoleAssignmentJson `json:"role"`
}

type RoleAssignmentJson struct {
	Id    string `json:"roleId"`
	Scope string `json:"scope"`
}

type ParentGroupsJson struct {
	Group []GroupJson `json:"group"`
}

type GroupDataModel struct {
	Id           types.String                   `tfsdk:"id"`
	Key          types.String                   `tfsdk:"key"`
	Name         types.String                   `tfsdk:"name"`
	Description  types.String                   `tfsdk:"description"`
	Roles        []RoleAssignmentGroupDataModel `tfsdk:"roles"`
	ParentGroups types.Set                      `tfsdk:"parent_groups"`
}

type RoleAssignmentGroupDataModel struct {
	Id      types.String `tfsdk:"id"`
	Global  types.Bool   `tfsdk:"global"`
	Project types.String `tfsdk:"project"`
}
