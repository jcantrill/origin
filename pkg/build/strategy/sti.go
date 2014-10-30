package strategy

import (
	"encoding/json"
	"io/ioutil"

	"github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	buildapi "github.com/openshift/origin/pkg/build/api"
)

// STIBuildStrategy creates STI(source to image) builds
type STIBuildStrategy struct {
	stiBuilderImage      string
	tempDirectoryCreator TempDirectoryCreator
	useLocalImage        bool
}

type TempDirectoryCreator interface {
	CreateTempDirectory() (string, error)
}

type tempDirectoryCreator struct{}

func (tc *tempDirectoryCreator) CreateTempDirectory() (string, error) {
	return ioutil.TempDir("", "stibuild")
}

var STITempDirectoryCreator = &tempDirectoryCreator{}

// NewSTIBuildStrategy creates a new STIBuildStrategy with the given
// builder image
func NewSTIBuildStrategy(stiBuilderImage string, tc TempDirectoryCreator, useLocalImage bool) *STIBuildStrategy {
	return &STIBuildStrategy{stiBuilderImage, tc, useLocalImage}
}

// CreateBuildPod creates a pod that will execute the STI build
// TODO: Make the Pod definition configurable
func (bs *STIBuildStrategy) CreateBuildPod(build *buildapi.Build) (*api.Pod, error) {
	buildJson, err := json.Marshal(build)
	if err != nil {
		return nil, err
	}
	pod := &api.Pod{
		JSONBase: api.JSONBase{
			ID: build.PodID,
		},
		DesiredState: api.PodState{
			Manifest: api.ContainerManifest{
				Version: "v1beta1",
				Containers: []api.Container{
					{
						Name:  "sti-build",
						Image: bs.stiBuilderImage,
						Env: []api.EnvVar{
							{Name: "BUILD_TAG", Value: build.Input.ImageTag},
							{Name: "SOURCE_URI", Value: build.Input.GitSource.URI},
							{Name: "SOURCE_REF", Value: build.Input.GitSource.Ref},
							{Name: "SOURCE_ID", Value: build.Input.GitSource.Commit.ID},
							{Name: "BUILD", Value: string(buildJson)},
							{Name: "REGISTRY", Value: build.Input.Registry},
							{Name: "BUILDER_IMAGE", Value: build.Input.STIInput.BuilderImage},
						},
					},
				},
				RestartPolicy: api.RestartPolicy{
					Never: &api.RestartPolicyNever{},
				},
			},
		},
	}
	if bs.useLocalImage {
		pod.DesiredState.Manifest.Containers[0].ImagePullPolicy = api.PullIfNotPresent
	}

	if err := bs.setupTempVolume(pod); err != nil {
		return nil, err
	}

	setupDockerSocket(pod)
	setupDockerConfig(pod)
	return pod, nil
}

func (bs *STIBuildStrategy) setupTempVolume(pod *api.Pod) error {
	tempDir, err := bs.tempDirectoryCreator.CreateTempDirectory()
	if err != nil {
		return err
	}
	tmpVolume := api.Volume{
		Name: "tmp",
		Source: &api.VolumeSource{
			HostDir: &api.HostDir{
				Path: tempDir,
			},
		},
	}
	tmpMount := api.VolumeMount{Name: "tmp", ReadOnly: false, MountPath: tempDir}
	pod.DesiredState.Manifest.Volumes = append(pod.DesiredState.Manifest.Volumes, tmpVolume)
	pod.DesiredState.Manifest.Containers[0].VolumeMounts =
		append(pod.DesiredState.Manifest.Containers[0].VolumeMounts, tmpMount)
	pod.DesiredState.Manifest.Containers[0].Env =
		append(pod.DesiredState.Manifest.Containers[0].Env, api.EnvVar{
			Name: "TEMP_DIR", Value: tempDir})

	return nil
}
