package ingressnginx

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	res, err := yamlv2.NewConfigFile(ctx, "ingress-nginx", &yamlv2.ConfigFileArgs{
		File: pulumi.String("https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v1.10.1/deploy/static/provider/cloud/deploy.yaml"),
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	_, err = corev1.NewConfigMapPatch(ctx, "ingress-nginx-controller", &corev1.ConfigMapPatchArgs{
		Metadata: &metav1.ObjectMetaPatchArgs{
			Name:      pulumi.String("ingress-nginx-controller"),
			Namespace: pulumi.String("ingress-nginx"),
			Annotations: pulumi.StringMap{
				"pulumi.com/patchForce": pulumi.String("true"),
			},
		},
		Data: pulumi.StringMap{
			"allow-snippet-annotations": pulumi.String("true"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{res}))
	if err != nil {
		return nil, err
	}

	resId := pulumi.Resource(res)

	return &resId, nil
}
