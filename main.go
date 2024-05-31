package main

import (
	"homelab-pulumi/kubernetes/ingressnginx"
	"homelab-pulumi/kubernetes/metallb"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Deploy MetalLB
		res, err := metallb.Deploy(ctx, nil)
		if err != nil {
			return err
		}

		// Deploy  Ingress-Nginx Controller
		res, err = ingressnginx.Deploy(ctx, res)
		if err != nil {
			return err
		}

		return nil
	})
}
