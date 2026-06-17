package teamcity

import (
	"terraform-provider-teamcity/models"
	"testing"
)

func strPtr(s string) *string { return &s }

// getProjectsAttrValue must drop virtual projects (auto-generated, inherited
// from a parent project) so they never enter Terraform state and cause drift.
func TestGetProjectsAttrValue_SkipsVirtual(t *testing.T) {
	data := []models.ProjectJson{
		{Name: "managed-a", Id: strPtr("Managed_A")},
		{Name: "auto-generated", Id: strPtr("Gradle_Master_Check_Quick_2_bucket1_virtual"), Virtual: true},
		{Name: "managed-b", Id: strPtr("Managed_B")},
		{Name: "auto-generated-stage", Id: strPtr("Gradle_Release7x_Check_Stage_PullRequestFeedback_X"), Virtual: true},
	}

	got := getProjectsAttrValue(data)

	if len(got) != 2 {
		t.Fatalf("expected 2 non-virtual projects, got %d: %v", len(got), got)
	}

	want := map[string]bool{
		`"Managed_A"`: true,
		`"Managed_B"`: true,
	}
	for _, v := range got {
		if !want[v.String()] {
			t.Errorf("unexpected project in result: %s", v.String())
		}
	}
}
