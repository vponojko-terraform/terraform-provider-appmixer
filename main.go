package main

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/plugin"

	// Import the internal package using the correct module path
	"github.com/vponojko-terraform/terraform-provider-appmixer/internal"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: func() *schema.Provider {
			// Use the Provider function from the internal package
			return internal.Provider()
		},
	})
}
