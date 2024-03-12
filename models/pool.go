package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
)

type PoolJson struct {
	Name string `json:"name"`
	Id   *int64 `json:"id,omitempty"`
	Size *int64 `json:"maxAgents,omitempty"`
}

type PoolDataModel struct {
	Name types.String `tfsdk:"name"`
	Id   types.Int64  `tfsdk:"id"`
	Size types.Int64  `tfsdk:"size"`
}
