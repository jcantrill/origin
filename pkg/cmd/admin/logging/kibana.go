package logging

import (
	"fmt"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"

	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

func (cfg *KibanaConfig) createKibana(deployName, publicMasterURL, imagePrefix, esHost string, esPort uint) []runtime.Object {
	// TODO version comes from where?
	image := fmt.Sprintf("%slogging-kibana:%s", imagePrefix, "version")
	proxyImage := fmt.Sprintf("%slogging-auth-proxy:%s", imagePrefix, "version")
	proxyName := fmt.Sprintf("%s-proxy", componentKibana)

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
				Type: deployapi.DeploymentStrategyTypeRolling,
				RollingParams: &deployapi.RollingDeploymentStrategyParams{
					IntervalSeconds:     int64Ptr(defaultDCIntervalSec),
					TimeoutSeconds:      int64Ptr(defaultDCTimeoutSec),
					UpdatePeriodSeconds: int64Ptr(defaultDCUpdatePeriodSec),
				},
			},
			Template: &kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{Name: deployName, Labels: labels},
				Spec: kapi.PodSpec{
					ServiceAccountName: fmt.Sprintf("%s-%s", namePrefixServiceAccount, componentKibana),
					Containers: []kapi.Container{
						{
							Name:            componentKibana,
							Image:           image,
							ImagePullPolicy: "Always",
							Env: []kapi.EnvVar{
								{Name: "ES_HOST", Value: esHost},
								{Name: "ES_PORT", Value: string(esPort)},
							},
							VolumeMounts: []kapi.VolumeMount{
								{Name: componentKibana,
									MountPath: "/etc/kibana/keys",
									ReadOnly:  true,
								},
							},
						},
						{
							Name:            proxyName,
							Image:           proxyImage,
							ImagePullPolicy: "Always",
							Ports: []kapi.ContainerPort{
								{Name: "oaproxy", ContainerPort: 3000},
							},
							Env: []kapi.EnvVar{
								{Name: "OAP_BACKEND_URL", Value: "http://localhost:5601"},
								{Name: "OAP_AUTH_MODE", Value: "oauth2"},
								{Name: "OAP_TRANSFORM", Value: "user_header,token_header"},
								{Name: "OAP_OAUTH_ID", Value: proxyName},
								{Name: "OAP_MASTER_URL", Value: defaultMasterURL},
								{Name: "OAP_PUBLIC_MASTER_URL", Value: publicMasterURL},
								{Name: "OAP_LOGOUT_REDIRECT", Value: fmt.Sprintf("%s/console/logout", publicMasterURL)},
								{Name: "OAP_MASTER_CA_FILE", Value: "/var/run/secrets/kubernetes.io/serviceaccount/ca.crt"},
								{Name: "OAP_DEBUG", Value: "false"},
							},
							VolumeMounts: []kapi.VolumeMount{
								{Name: proxyName,
									MountPath: "/secret",
									ReadOnly:  true,
								},
							},
						},
					},
					Volumes: []kapi.Volume{
						kapi.Volume{Name: componentKibana, VolumeSource: kapi.VolumeSource{Secret: &kapi.SecretVolumeSource{SecretName: "logging-kibana"}}},
						kapi.Volume{Name: proxyName, VolumeSource: kapi.VolumeSource{Secret: &kapi.SecretVolumeSource{SecretName: "logging-kibana-proxy"}}},
					},
				},
			},
		},
	}
	return []runtime.Object{dc}
}
