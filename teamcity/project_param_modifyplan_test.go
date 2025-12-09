package teamcity

import (
	"github.com/hashicorp/terraform-plugin-framework/types"
	"terraform-provider-teamcity/models"
	"testing"
)

func TestShouldWarnReplace_True_ForPassword(t *testing.T) {
	plan := paramResourceModel{
		ProjectId: types.StringValue("_Root"),
		Name:      types.StringValue("secret_token"),
		Type:      types.StringValue(models.ParamTypePassword),
	}
	if !isSecureParam(plan) {
		t.Fatalf("expected isSecureParam to be true for password type")
	}
}

func TestShouldWarnReplace_False_ForNonPassword(t *testing.T) {
	plan := paramResourceModel{
		ProjectId: types.StringValue("_Root"),
		Name:      types.StringValue("param"),
		Type:      types.StringValue(models.ParamTypeText),
	}
	if isSecureParam(plan) {
		t.Fatalf("expected isSecureParam to be false for non-password type")
	}
}
