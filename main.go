package main

import (
	"homelab-pulumi/k8s/authentik"
	"homelab-pulumi/k8s/certmanager"
	"homelab-pulumi/k8s/democraticcsi"
	"homelab-pulumi/k8s/ingressnginx"
	"homelab-pulumi/k8s/metallb"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Deploy MetalLB
		res, err := metallb.Deploy(ctx, nil)
		if err != nil {
			return err
		}

		// Deploy Ingress-Nginx controller
		res, err = ingressnginx.Deploy(ctx, res)
		if err != nil {
			return err
		}

		// Deploy Democratic CSI driver
		res, err = democraticcsi.Deploy(ctx, res)
		if err != nil {
			return err
		}

		// Deploy Cert-Manager
		res, err = certmanager.Deploy(ctx, res)
		if err != nil {
			return err
		}

		// Deploy Authentik
		res, err = authentik.Deploy(ctx, res)
		if err != nil {
			return err
		}

		return nil
	})
}
