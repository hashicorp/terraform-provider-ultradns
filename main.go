package main

import (
	"github.com/hashicorp/terraform/plugin"
	"github.com/manheim/terraform-provider-ultradns/ultradns"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: ultradns.Provider})
}
