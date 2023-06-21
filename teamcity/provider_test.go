package teamcity

import (
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

const (
	providerConfig = `
terraform {
  required_providers {
    teamcity = {
      source = "jetbrains/teamcity"
    }
  }
}
provider "teamcity" {
#	host = "http://localhost:8111"
}
`
)

var (
	testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
		"teamcity": providerserver.NewProtocol6WithError(New()),
	}
)
