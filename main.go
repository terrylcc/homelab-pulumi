package main

import (
	"homelab-pulumi/kubernetes/ingressnginx"
	"homelab-pulumi/kubernetes/metallb"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Deploy MetalLB
		err := metallb.Deploy(ctx)
		if err != nil {
			return err
		}

		// Deploy  Ingress-Nginx Controller
		err = ingressnginx.Deploy(ctx)
		if err != nil {
			return err
		}

		return nil
	})
}
