package teamcity

import (
	"terraform-provider-teamcity/models"
	"testing"
)

func TestConvertToResource_NilParent(t *testing.T) {
	r := &projectResource{}
	id := "test-id"
	project := models.ProjectJson{
		Name:          "test-name",
		Id:            &id,
		ParentProject: nil,
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("The code panicked: %v", r)
		}
	}()

	res := r.convertToResource(project)
	if res.ParentProjectId.ValueString() != defaultParentProjectId {
		t.Errorf("Expected ParentProjectId to be %s, got %s", defaultParentProjectId, res.ParentProjectId.ValueString())
	}
}

func TestConvertToResource_Full(t *testing.T) {
	r := &projectResource{}
	id := "test-id"
	parentId := "parent-id"
	project := models.ProjectJson{
		Name: "test-name",
		Id:   &id,
		ParentProject: &models.ProjectJson{
			Id: &parentId,
		},
	}

	res := r.convertToResource(project)
	if res.Name.ValueString() != "test-name" {
		t.Errorf("Expected Name to be test-name, got %s", res.Name.ValueString())
	}
	if res.Id.ValueString() != id {
		t.Errorf("Expected Id to be %s, got %s", id, res.Id.ValueString())
	}
	if res.ParentProjectId.ValueString() != parentId {
		t.Errorf("Expected ParentProjectId to be %s, got %s", parentId, res.ParentProjectId.ValueString())
	}
}
