package logging

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

func (cfg *CuratorConfig) createCurator(deployName, imagePrefix, esHost string, esPort uint) []runtime.Object {
	// TODO version comes from where?
	image := fmt.Sprintf("%slogging-curator:%s", imagePrefix, "version")

	labels := labels.Set(map[string]string{
		"provider":  "openshift",
		"component": deployName,
	})
	dc := &deployapi.DeploymentConfig{
		ObjectMeta: kapi.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", namePrefixDeploymentConfig, deployName),
			Labels: labels,
		},
		Spec: deployapi.DeploymentConfigSpec{
			Replicas: 1,
			Selector: labels,
			Strategy: deployapi.DeploymentStrategy{
				Type: deployapi.DeploymentStrategyTypeRecreate,
				RecreateParams: &deployapi.RecreateDeploymentStrategyParams{
					TimeoutSeconds: int64Ptr(defaultDCTimeoutSec),
				},
			},
			Template: &kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{Name: deployName, Labels: labels},
				Spec: kapi.PodSpec{
					ServiceAccountName: fmt.Sprintf("%s-%s", namePrefixServiceAccount, componentCurator),
					Containers: []kapi.Container{
						{
							Name:            componentCurator,
							Image:           image,
							ImagePullPolicy: "Always",
							Resources: kapi.ResourceRequirements{
								Limits: map[kapi.ResourceName]resource.Quantity{
									kapi.ResourceCPU: resource.MustParse(curatorCPU),
								},
							},
							Env: []kapi.EnvVar{
								{Name: "ES_HOST", Value: esHost},
								{Name: "ES_PORT", Value: string(esPort)},
								{Name: "K8S_HOST_URL", Value: defaultMasterURL},
								{Name: "ES_CLIENT_CERT", Value: "/etc/curator/keys/cert"},
								{Name: "ES_CLIENT_KEY", Value: "/etc/curator/keys/key"},
								{Name: "ES_CA", Value: "/etc/curator/keys/ca"},
								{Name: "CURATOR_SCRIPT_LOG_LEVEL", Value: "INFO"},
								{Name: "CURATOR_LOG_LEVEL", Value: "ERROR"},
							},
							VolumeMounts: []kapi.VolumeMount{
								{Name: "certs",
									MountPath: "/etc/curator/keys",
									ReadOnly:  true,
								},
								{Name: "config",
									MountPath: "/etc/curator/settings",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []kapi.Volume{
						kapi.Volume{Name: "certs", VolumeSource: kapi.VolumeSource{Secret: &kapi.SecretVolumeSource{SecretName: deployName}}},
						kapi.Volume{
							Name: "config",
							VolumeSource: kapi.VolumeSource{
								ConfigMap: &kapi.ConfigMapVolumeSource{
									LocalObjectReference: kapi.LocalObjectReference{Name: deployName},
								}}},
					},
				},
			},
		},
	}
	return []runtime.Object{dc}

}
