package authentik

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "authentik")

	postgresPassword := cfg.RequireSecret("postgresPassword")

	secretKey := cfg.RequireSecret("secretKey")

	ns, err := corev1.NewNamespace(ctx, "authentik", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("authentik"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	res, err := helmv3.NewRelease(ctx, "authentik", &helmv3.ReleaseArgs{
		Namespace: ns.Metadata.Name(),
		Chart:     pulumi.String("authentik"),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String("https://charts.goauthentik.io"),
		},
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset("assets/authentik/values.yaml"),
		},
		Values: pulumi.Map{
			"authentik": pulumi.Map{
				"secret_key": secretKey,
				"postgresql": pulumi.Map{
					"password": postgresPassword,
				},
			},
			"postgresql": pulumi.Map{
				"auth": pulumi.Map{
					"password": postgresPassword,
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
