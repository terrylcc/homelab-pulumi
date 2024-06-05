package certmanager

import (
	apiextensions "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apiextensions"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	yamlv2 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/yaml/v2"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "")

	email := cfg.Require("email")

	apiToken := cfg.RequireSecret("apiToken")

	res, err := yamlv2.NewConfigFile(ctx, "cert-manager", &yamlv2.ConfigFileArgs{
		File: pulumi.String("https://github.com/cert-manager/cert-manager/releases/download/v1.14.5/cert-manager.yaml"),
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	sec, err := corev1.NewSecret(ctx, "cloudflare-api-token-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("cloudflare-api-token-secret"),
			Namespace: pulumi.String("cert-manager"),
		},
		Type: pulumi.String("Opaque"),
		StringData: pulumi.StringMap{
			"api-token": apiToken,
		},
	}, pulumi.DependsOn([]pulumi.Resource{res}))
	if err != nil {
		return nil, err
	}

	_, err = apiextensions.NewCustomResource(ctx, "letsencrypt-staging", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("letsencrypt-staging"),
		},
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"acme": map[string]interface{}{
					"email":  email,
					"server": pulumi.String("https://acme-staging-v02.api.letsencrypt.org/directory"),
					"privateKeySecretRef": map[string]interface{}{
						"name": pulumi.String("letsencrypt-staging-private-key"),
					},
					"solvers": []interface{}{
						map[string]interface{}{
							"dns01": map[string]interface{}{
								"cloudflare": map[string]interface{}{
									"apiTokenSecretRef": map[string]interface{}{
										"name": pulumi.String("cloudflare-api-token-secret"),
										"key":  pulumi.String("api-token"),
									},
								},
							},
						},
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{res, sec}))
	if err != nil {
		return nil, err
	}

	_, err = apiextensions.NewCustomResource(ctx, "letsencrypt-prod", &apiextensions.CustomResourceArgs{
		ApiVersion: pulumi.String("cert-manager.io/v1"),
		Kind:       pulumi.String("ClusterIssuer"),
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("letsencrypt-prod"),
		},
		OtherFields: map[string]interface{}{
			"spec": map[string]interface{}{
				"acme": map[string]interface{}{
					"email":  email,
					"server": pulumi.String("https://acme-v02.api.letsencrypt.org/directory"),
					"privateKeySecretRef": map[string]interface{}{
						"name": pulumi.String("letsencrypt-prod-private-key"),
					},
					"solvers": []interface{}{
						map[string]interface{}{
							"dns01": map[string]interface{}{
								"cloudflare": map[string]interface{}{
									"apiTokenSecretRef": map[string]interface{}{
										"name": pulumi.String("cloudflare-api-token-secret"),
										"key":  pulumi.String("api-token"),
									},
								},
							},
						},
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{res, sec}))
	if err != nil {
		return nil, err
	}

	resId := pulumi.Resource(res)

	return &resId, nil
}
