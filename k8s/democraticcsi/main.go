package democraticcsi

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "democraticcsi")

	privateKey := cfg.RequireSecret("privateKey")

	ns, err := corev1.NewNamespace(ctx, "democratic-csi", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("democratic-csi"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	res, err := helmv3.NewRelease(ctx, "democratic-csi", &helmv3.ReleaseArgs{
		Namespace: pulumi.String("democratic-csi"),
		Chart:     pulumi.String("democratic-csi"),
		RepositoryOpts: &helmv3.RepositoryOptsArgs{
			Repo: pulumi.String("https://democratic-csi.github.io/charts"),
		},
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset("./assets/democraticcsi/values.yaml"),
		},
		Values: pulumi.Map{
			"driver": pulumi.Map{
				"config": pulumi.Map{
					"sshConnection": pulumi.Map{
						"privateKey": privateKey,
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
