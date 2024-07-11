package kubeprometheusstack

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "kubeprometheusstack")

	oauthClientId := cfg.Require("oauthClientId")

	oauthClientSecret := cfg.RequireSecret("oauthClientSecret")

	ns, err := corev1.NewNamespace(ctx, "kube-prometheus-stack", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("kube-prometheus-stack"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	res, err := helmv3.NewRelease(ctx, "kube-prometheus-stack", &helmv3.ReleaseArgs{
		Namespace: ns.Metadata.Name(),
		Chart:     pulumi.String("kube-prometheus-stack"),
		Version:   pulumi.String("61.3.0"),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String("https://prometheus-community.github.io/helm-charts"),
		},
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset("assets/kubeprometheusstack/values.yaml"),
		},
		Values: pulumi.Map{
			"grafana": pulumi.Map{
				"grafana.ini": pulumi.Map{
					"auth.generic_oauth": pulumi.Map{
						"client_id":     pulumi.String(oauthClientId),
						"client_secret": oauthClientSecret,
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	resId := pulumi.Resource(res)

	return &resId, nil
}
