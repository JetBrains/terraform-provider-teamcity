package main

import (
	"context"
	"flag"
	"log"
	"terraform-provider-teamcity/teamcity"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
)

var (
	version string = "dev"
)

func main() {
	var debug bool

	flag.BoolVar(&debug, "debug", false, "set to true to run the provider with support for debuggers like delve")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/jetbrains/teamcity",
		Debug:   debug,
	}

	err := providerserver.Serve(context.Background(), teamcity.New(version), opts)
	if err != nil {
		log.Fatal(err.Error())
	}
}
