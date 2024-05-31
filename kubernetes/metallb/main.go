package metallb

import (
	apiextensions "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func Deploy(ctx *pulumi.Context) error {
	stack, err := yamlv2.NewConfigFile(ctx, "metallb-manifest", &yamlv2.ConfigFileArgs{
		File: pulumi.String("https://raw.githubusercontent.com/metallb/metallb/v0.14.5/config/manifests/metallb-native.yaml"),
	})
	if err != nil {
		return err
	}

	_, err = apiextensions.NewCustomResource(ctx, "metallb-ipaddresspool", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("metallb.io/v1beta1"),
		Kind:       pulumi.String("IPAddressPool"),
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: pulumi.String("metallb-system"),
			Name:      pulumi.String("metallb-ipaddresspool"),
		},
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"addresses": []interface{}{
					"10.9.0.0/24",
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{stack}))
	if err != nil {
		return err
	}

	_, err = apiextensions.NewCustomResource(ctx, "metallb-l2advertisement", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("metallb.io/v1beta1"),
		Kind:       pulumi.String("L2Advertisement"),
		Metadata: &metav1.ObjectMetaArgs{
			Namespace: pulumi.String("metallb-system"),
			Name:      pulumi.String("metallb-l2advertisement"),
		},
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"ipAddressPools": []interface{}{
					"metallb-ipaddresspool",
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{stack}))
	if err != nil {
		return err
	}

	return nil
}
