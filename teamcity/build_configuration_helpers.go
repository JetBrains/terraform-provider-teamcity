package teamcity

import (
	"terraform-provider-teamcity/models"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// mergePropertiesFromServer maps a TeamCity properties payload into a Terraform
// state value while preserving the user-intended shape of an absent map.
//
// When the server returns no properties, we cannot blindly write MapNull: if the
// user wrote `properties = {}` in HCL, the prior state holds a non-null empty
// map, and replacing it with MapNull produces a perpetual `+ properties = {}`
// plan diff (null != empty in Terraform's diff engine). Conversely, if the user
// omitted `properties` entirely the prior state is null and we keep it that way.
//
// Used by build_configuration_feature, _step, and _trigger resources, which all
// expose a generic Optional+Computed `properties` map attribute.
func mergePropertiesFromServer(actual *models.Properties, current types.Map, diags *diag.Diagnostics) types.Map {
	if actual == nil || len(actual.Property) == 0 {
		if current.IsNull() || current.IsUnknown() {
			return types.MapNull(types.StringType)
		}
		return types.MapValueMust(types.StringType, map[string]attr.Value{})
	}

	propsMap := make(map[string]attr.Value, len(actual.Property))
	for _, p := range actual.Property {
		propsMap[p.Name] = types.StringValue(p.Value)
	}
	props, d := types.MapValue(types.StringType, propsMap)
	diags.Append(d...)
	if d.HasError() {
		return current
	}
	return props
}
