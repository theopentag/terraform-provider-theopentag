package main

import (
	"context"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/theopentag/terraform-provider-theopentag/internal/provider"
)

var version string

func main() {
	err := providerserver.Serve(context.Background(), provider.New, providerserver.ServeOpts{
		Address: "registry.terraform.io/theopentag/theopentag",
	})
	if err != nil {
		log.Fatal(err)
	}
}
