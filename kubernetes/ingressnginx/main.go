package ingressnginx

import (
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Deploy(ctx *pulumi.Context) error {
	_, err := yamlv2.NewConfigFile(ctx, "ingressnginx-manifest", &yamlv2.ConfigFileArgs{
		File: pulumi.String("https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.1/deploy/static/provider/cloud/deploy.yaml"),
	})
	if err != nil {
		return err
	}

	return nil
}
