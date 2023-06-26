package client

type BuildTypes struct {
	BuildType []BuildType `json:"buildType"`
}

type BuildType struct {
	Id string `json:"id"`
}
