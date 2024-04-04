package models

type Properties struct {
	Property []Property `json:"property"`
}
type Property struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}
