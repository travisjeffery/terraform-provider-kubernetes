package main

import (
	"github.com/confluentinc/terraform-provider-kubernetes/kubernetes"
	"github.com/hashicorp/terraform/plugin"
)

func main() {
	plugin.Serve(&plugin.ServeOpts{
		ProviderFunc: kubernetes.Provider})
}
