package main

import (
	"context"
	"terraform-provider-teamcity/teamcity"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

func main() {
	providerserver.Serve(context.Background(), teamcity.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/jetbrains/teamcity",
	})
}
