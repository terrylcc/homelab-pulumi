package ocis

import (
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	helmv3 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/helm/v3"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "ocis")

	adminUuid := cfg.Require("adminUuid")

	ldapUri := cfg.Require("ldapUri")

	webClientId := cfg.Require("webClientId")

	jwtSecret := cfg.RequireSecret("jwtSecret")

	machineAuthApiKey := cfg.RequireSecret("machineAuthApiKey")

	storageSystemUserId := cfg.RequireSecret("storageSystemUserId")

	storageSystemApiKey := cfg.RequireSecret("storageSystemApiKey")

	storageSystemJwtSecret := cfg.RequireSecret("storageSystemJwtSecret")

	transferSecret := cfg.RequireSecret("transferSecret")

	thumbnailsTransferSecret := cfg.RequireSecret("thumbnailsTransferSecret")

	ldapBindPassword := cfg.RequireSecret("ldapBindPassword")

	storageUuid := cfg.Require("storageUuid")

	applicationId := cfg.Require("applicationId")

	ns, err := corev1.NewNamespace(ctx, "ocis", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("ocis"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	jwtSec, err := corev1.NewSecret(ctx, "jwt-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("jwt-secret"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"jwt-secret": jwtSecret,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	machineAuthApiKeySec, err := corev1.NewSecret(ctx, "machine-auth-api-key", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("machine-auth-api-key"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"machine-auth-api-key": machineAuthApiKey,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	storageSystemSec, err := corev1.NewSecret(ctx, "storage-system", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("storage-system"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"user-id": storageSystemUserId,
			"api-key": storageSystemApiKey,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	storageSystemJwtSec, err := corev1.NewSecret(ctx, "storage-system-jwt-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("storage-system-jwt-secret"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"storage-system-jwt-secret": storageSystemJwtSecret,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	transferSec, err := corev1.NewSecret(ctx, "transfer-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("transfer-secret"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"transfer-secret": transferSecret,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	thumbnailsTransferSec, err := corev1.NewSecret(ctx, "thumbnails-transfer-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("thumbnails-transfer-secret"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"thumbnails-transfer-secret": thumbnailsTransferSecret,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	ldapBindSec, err := corev1.NewSecret(ctx, "ldap-bind-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("ldap-bind-secret"),
			Namespace: pulumi.String("ocis"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"reva-ldap-bind-password": ldapBindPassword,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	storageUsersCfg, err := corev1.NewConfigMap(ctx, "storage-users", &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("storage-users"),
			Namespace: pulumi.String("ocis"),
		},
		Data: pulumi.StringMap{
			"storage-uuid": pulumi.String(storageUuid),
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	graphCfg, err := corev1.NewConfigMap(ctx, "graph", &corev1.ConfigMapArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("graph"),
			Namespace: pulumi.String("ocis"),
		},
		Data: pulumi.StringMap{
			"application-id": pulumi.String(applicationId),
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	res, err := helmv3.NewRelease(ctx, "ocis", &helmv3.ReleaseArgs{
		Chart:     pulumi.String("assets/ocis/ocis-charts/charts/ocis"),
		Namespace: ns.Metadata.Name(),
		ValueYamlFiles: pulumi.AssetOrArchiveArray{
			pulumi.NewFileAsset("assets/ocis/values.yaml"),
		},
		Values: pulumi.Map{
			"features": pulumi.Map{
				"externalUserManagement": pulumi.Map{
					"adminUUID": pulumi.String(adminUuid),
					"oidc": pulumi.Map{
						"webClientID": pulumi.String(webClientId),
					},
					"ldap": pulumi.Map{
						"uri": pulumi.String(ldapUri),
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{
		ns,
		jwtSec,
		machineAuthApiKeySec,
		storageSystemSec,
		storageSystemJwtSec,
		transferSec,
		thumbnailsTransferSec,
		ldapBindSec,
		storageUsersCfg,
		graphCfg,
	}))
	if err != nil {
		return nil, err
	}

	resId := pulumi.Resource(res)

	return &resId, nil
}
