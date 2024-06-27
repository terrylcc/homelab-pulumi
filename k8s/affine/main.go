package affine

import (
	"encoding/base64"
	"fmt"

	appsv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/apps/v1"
	corev1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/core/v1"
	metav1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/meta/v1"
	networkingv1 "github.com/pulumi/pulumi-kubernetes/sdk/v4/go/kubernetes/networking/v1"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func Deploy(ctx *pulumi.Context, id *pulumi.Resource) (*pulumi.Resource, error) {
	cfg := config.New(ctx, "affine")

	postgresPassword := cfg.RequireSecret("postgresPassword")

	affineEmail := cfg.RequireSecret("affineEmail")

	affinePassword := cfg.RequireSecret("affinePassword")

	ns, err := corev1.NewNamespace(ctx, "affine", &corev1.NamespaceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name: pulumi.String("affine"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{*id}))
	if err != nil {
		return nil, err
	}

	postgresSec, err := corev1.NewSecret(ctx, "postgres-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("postgres-secret"),
			Namespace: pulumi.String("affine"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"password": postgresPassword,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	postgresData, err := corev1.NewPersistentVolumeClaim(ctx, "postgres-data", &corev1.PersistentVolumeClaimArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("postgres-data"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.PersistentVolumeClaimSpecArgs{
			AccessModes: pulumi.StringArray{
				pulumi.String("ReadWriteOnce"),
			},
			Resources: &corev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String("2Ti"),
				},
			},
			StorageClassName: pulumi.String("zfs-iscsi-csi"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	postgresDeploy, err := appsv1.NewDeployment(ctx, "postgres", &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("postgres"),
			Namespace: pulumi.String("affine"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("postgres"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("postgres"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("postgres"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("postgres"),
							Image: pulumi.String("postgres"),
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("POSTGRES_USER"),
									Value: pulumi.String("affine"),
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("POSTGRES_PASSWORD"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											Key:  pulumi.String("password"),
											Name: pulumi.String("postgres-secret"),
										},
									},
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("POSTGRES_DB"),
									Value: pulumi.String("affine"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("PGDATA"),
									Value: pulumi.String("/var/lib/postgresql/data/pgdata"),
								},
							},
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(5432),
								},
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									MountPath: pulumi.String("/var/lib/postgresql/data"),
									Name:      pulumi.String("postgres-data"),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								Exec: &corev1.ExecActionArgs{
									Command: pulumi.StringArray{
										pulumi.String("pg_isready"),
										pulumi.String("--dbname"),
										pulumi.String("affine"),
									},
								},
							},
						},
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("postgres-data"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pulumi.String("postgres-data"),
							},
						},
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, postgresSec, postgresData}))
	if err != nil {
		return nil, err
	}

	postgresSvc, err := corev1.NewService(ctx, "postgres", &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("postgres"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Name:     pulumi.String("postgres"),
					Port:     pulumi.Int(5432),
					Protocol: pulumi.String("TCP"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("postgres"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, postgresDeploy}))
	if err != nil {
		return nil, err
	}

	redisData, err := corev1.NewPersistentVolumeClaim(ctx, "redis-data", &corev1.PersistentVolumeClaimArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("redis-data"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.PersistentVolumeClaimSpecArgs{
			AccessModes: pulumi.StringArray{
				pulumi.String("ReadWriteOnce"),
			},
			Resources: &corev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String("8Gi"),
				},
			},
			StorageClassName: pulumi.String("zfs-iscsi-csi"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	redisDeploy, err := appsv1.NewDeployment(ctx, "redis", &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("redis"),
			Namespace: pulumi.String("affine"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("redis"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("redis"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("redis"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("redis"),
							Image: pulumi.String("redis"),
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(6379),
								},
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									MountPath: pulumi.String("/data"),
									Name:      pulumi.String("redis-data"),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								Exec: &corev1.ExecActionArgs{
									Command: pulumi.StringArray{
										pulumi.String("redis-cli"),
										pulumi.String("--raw"),
										pulumi.String("incr"),
										pulumi.String("ping"),
									},
								},
							},
						},
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("redis-data"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pulumi.String("redis-data"),
							},
						},
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, redisData}))
	if err != nil {
		return nil, err
	}

	redisSvc, err := corev1.NewService(ctx, "redis", &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("redis"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Name:     pulumi.String("redis"),
					Port:     pulumi.Int(6379),
					Protocol: pulumi.String("TCP"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("redis"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, redisDeploy}))
	if err != nil {
		return nil, err
	}

	affineSec, err := corev1.NewSecret(ctx, "affine-secret", &corev1.SecretArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("affine-secret"),
			Namespace: pulumi.String("affine"),
		},
		Type: pulumi.String("Opaque"),
		Data: pulumi.StringMap{
			"password": affinePassword,
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	affineConfig, err := corev1.NewPersistentVolumeClaim(ctx, "affine-config", &corev1.PersistentVolumeClaimArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("affine-config"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.PersistentVolumeClaimSpecArgs{
			AccessModes: pulumi.StringArray{
				pulumi.String("ReadWriteOnce"),
			},
			Resources: &corev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String("8Gi"),
				},
			},
			StorageClassName: pulumi.String("zfs-iscsi-csi"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	affineData, err := corev1.NewPersistentVolumeClaim(ctx, "affine-data", &corev1.PersistentVolumeClaimArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("affine-data"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.PersistentVolumeClaimSpecArgs{
			AccessModes: pulumi.StringArray{
				pulumi.String("ReadWriteOnce"),
			},
			Resources: &corev1.VolumeResourceRequirementsArgs{
				Requests: pulumi.StringMap{
					"storage": pulumi.String("8Gi"),
				},
			},
			StorageClassName: pulumi.String("zfs-iscsi-csi"),
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns}))
	if err != nil {
		return nil, err
	}

	affineDeploy, err := appsv1.NewDeployment(ctx, "affine", &appsv1.DeploymentArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("affine"),
			Namespace: pulumi.String("affine"),
			Labels: pulumi.StringMap{
				"app": pulumi.String("affine"),
			},
		},
		Spec: &appsv1.DeploymentSpecArgs{
			Replicas: pulumi.Int(1),
			Selector: &metav1.LabelSelectorArgs{
				MatchLabels: pulumi.StringMap{
					"app": pulumi.String("affine"),
				},
			},
			Template: &corev1.PodTemplateSpecArgs{
				Metadata: &metav1.ObjectMetaArgs{
					Labels: pulumi.StringMap{
						"app": pulumi.String("affine"),
					},
				},
				Spec: &corev1.PodSpecArgs{
					Containers: corev1.ContainerArray{
						&corev1.ContainerArgs{
							Name:  pulumi.String("affine"),
							Image: pulumi.String("ghcr.io/toeverything/affine-graphql:stable"),
							Command: pulumi.StringArray{
								pulumi.String("/bin/bash"),
								pulumi.String("-c"),
								pulumi.String("node ./scripts/self-host-predeploy && node ./dist/index.js"),
							},
							Env: corev1.EnvVarArray{
								&corev1.EnvVarArgs{
									Name:  pulumi.String("AFFINE_ADMIN_EMAIL"),
									Value: affineEmail,
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("AFFINE_ADMIN_PASSWORD"),
									ValueFrom: &corev1.EnvVarSourceArgs{
										SecretKeyRef: &corev1.SecretKeySelectorArgs{
											Key:  pulumi.String("password"),
											Name: pulumi.String("affine-secret"),
										},
									},
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("AFFINE_CONFIG_PATH"),
									Value: pulumi.String("/root/.affine/config"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("NODE_ENV"),
									Value: pulumi.String("production"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("NODE_OPTIONS"),
									Value: pulumi.String("--import=./scripts/register.js"),
								},
								&corev1.EnvVarArgs{
									Name: pulumi.String("DATABASE_URL"),
									Value: postgresPassword.ApplyT(func(pass string) string {
										data, _ := base64.StdEncoding.DecodeString(pass)
										return fmt.Sprintf("postgres://affine:%s@postgres.affine.svc.cluster.local:5432/affine", data)
									}).(pulumi.StringOutput),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("REDIS_SERVER_HOST"),
									Value: pulumi.String("redis.affine.svc.cluster.local"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("REDIS_SERVER_PORT"),
									Value: pulumi.String("6379"),
								},
								&corev1.EnvVarArgs{
									Name:  pulumi.String("TELEMETRY_ENABLE"),
									Value: pulumi.String("false"),
								},
							},
							Ports: corev1.ContainerPortArray{
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(3010),
								},
								&corev1.ContainerPortArgs{
									ContainerPort: pulumi.Int(5555),
								},
							},
							VolumeMounts: corev1.VolumeMountArray{
								&corev1.VolumeMountArgs{
									MountPath: pulumi.String("/root/.affine/config"),
									Name:      pulumi.String("affine-config"),
								},
								&corev1.VolumeMountArgs{
									MountPath: pulumi.String("/root/.affine/storage"),
									Name:      pulumi.String("affine-data"),
								},
							},
							LivenessProbe: &corev1.ProbeArgs{
								HttpGet: &corev1.HTTPGetActionArgs{
									Port: pulumi.Int(3010),
								},
								InitialDelaySeconds: pulumi.Int(180),
							},
						},
					},
					Volumes: corev1.VolumeArray{
						&corev1.VolumeArgs{
							Name: pulumi.String("affine-config"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pulumi.String("affine-config"),
							},
						},
						&corev1.VolumeArgs{
							Name: pulumi.String("affine-data"),
							PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSourceArgs{
								ClaimName: pulumi.String("affine-data"),
							},
						},
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, affineSec, affineConfig, affineData, postgresSvc, redisSvc}))
	if err != nil {
		return nil, err
	}

	affineSvc, err := corev1.NewService(ctx, "affine", &corev1.ServiceArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("affine"),
			Namespace: pulumi.String("affine"),
		},
		Spec: &corev1.ServiceSpecArgs{
			Ports: corev1.ServicePortArray{
				&corev1.ServicePortArgs{
					Name:     pulumi.String("http"),
					Port:     pulumi.Int(3010),
					Protocol: pulumi.String("TCP"),
				},
			},
			Selector: pulumi.StringMap{
				"app": pulumi.String("affine"),
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, affineDeploy}))
	if err != nil {
		return nil, err
	}

	affineIngress, err := networkingv1.NewIngress(ctx, "affine", &networkingv1.IngressArgs{
		Metadata: &metav1.ObjectMetaArgs{
			Name:      pulumi.String("affine"),
			Namespace: pulumi.String("affine"),
			Annotations: pulumi.StringMap{
				"cert-manager.io/cluster-issuer": pulumi.String("letsencrypt-prod"),
			},
		},
		Spec: &networkingv1.IngressSpecArgs{
			IngressClassName: pulumi.String("nginx"),
			Rules: networkingv1.IngressRuleArray{
				&networkingv1.IngressRuleArgs{
					Host: pulumi.String("affine.terri.cc"),
					Http: &networkingv1.HTTPIngressRuleValueArgs{
						Paths: networkingv1.HTTPIngressPathArray{
							&networkingv1.HTTPIngressPathArgs{
								Path:     pulumi.String("/"),
								PathType: pulumi.String("Prefix"),
								Backend: &networkingv1.IngressBackendArgs{
									Service: &networkingv1.IngressServiceBackendArgs{
										Name: pulumi.String("affine"),
										Port: &networkingv1.ServiceBackendPortArgs{
											Number: pulumi.Int(3010),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}, pulumi.DependsOn([]pulumi.Resource{ns, affineSvc}))
	if err != nil {
		return nil, err
	}

	resId := pulumi.Resource(affineIngress)

	return &resId, nil
}
