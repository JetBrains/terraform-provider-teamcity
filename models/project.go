package models

type Project struct {
	Name string  `json:"name"`
	Id   *string `json:"id,omitempty"`
}

type ProjectsJson struct {
	Project []Project `json:"project,omitempty"`
}
