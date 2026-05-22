package models

import "github.com/hashicorp/terraform-plugin-framework/types"

type VcsRootEntryJson struct {
	ID            string       `json:"id,omitempty"`
	VcsRoot       *VcsRootJson `json:"vcs-root,omitempty"`
	CheckoutRules string       `json:"checkout-rules,omitempty"`
}

type VcsRootEntryDataModel struct {
	ID                   types.String `tfsdk:"id"`
	BuildConfigurationId types.String `tfsdk:"build_configuration_id"`
	VcsRootId            types.String `tfsdk:"vcs_root_id"`
	CheckoutRules        types.String `tfsdk:"checkout_rules"`
}
