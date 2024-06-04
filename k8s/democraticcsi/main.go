package democraticcsi

import (
	helmv4 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v4"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "")

	privateKey := cfg.RequireSecret("privateKey")

	res, err := helmv4.NewChart(ctx, "democratic-csi-chart", &helmv4.ChartArgs{
		Chart: pulumi.String("democratic-csi"),
		RepositoryOpts: &helmv4.RepositoryOptsArgs{
			Repo: pulumi.String("https://democratic-csi.github.io/charts"),
		},
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset("./k8s/democraticcsi/values.yaml"),
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
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	resId := pulumi.Resource(res)

	return &resId, nil
}
