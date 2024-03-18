package models

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-framework/types/basetypes"
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

func (p *PoolJson) GetSize () types.Int64 {
    if p.Size == nil {
        return basetypes.NewInt64Null()
    } else {
        return types.Int64Value(int64(*(p.Size)))
    }
}
