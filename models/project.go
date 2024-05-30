package models

type ProjectJson struct {
	Name string  `json:"name"`
	Id   *string `json:"id,omitempty"`
}

type ProjectsJson struct {
	Project []ProjectJson `json:"project,omitempty"`
}
