// Command terraform-provider-vault-secrets-ghp is the Terraform/OpenTofu
// provider for the vault-plugin-secrets-ghp secrets engine: it manages the
// engine's mount + connection config and its token-minting roles on HashiCorp
// Vault or OpenBao.
package main

//go:generate go run github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.21.0 generate --provider-name ghp

import (
	"context"
	"flag"
	"log"

	"github.com/hashicorp/terraform-plugin-framework/providerserver"

	"github.com/IUCLabs/terraform-provider-vault-secrets-ghp/internal/provider"
)

var version = "0.0.1"

func main() {
	var debug bool
	flag.BoolVar(&debug, "debug", false, "enable debug mode")
	flag.Parse()

	opts := providerserver.ServeOpts{
		Address: "registry.terraform.io/iuclabs/vault-secrets-ghp",
		Debug:   debug,
	}

	if err := providerserver.Serve(context.Background(), provider.New(version), opts); err != nil {
		log.Fatal(err)
	}
}
