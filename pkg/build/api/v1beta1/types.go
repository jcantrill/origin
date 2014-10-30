package v1beta1

import (
	api "github.com/GoogleCloudPlatform/kubernetes/pkg/api/v1beta1"
)

// Build encapsulates the inputs needed to produce a new deployable image, as well as
// the status of the operation and a reference to the Pod which runs the build.
type Build struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// Input is the set of inputs used to configure the build
	Input BuildInput `json:"input,omitempty" yaml:"input,omitempty"`

	// Status is the current status of the build
	Status BuildStatus `json:"status,omitempty" yaml:"status,omitempty"`

	// PodID is the id of the pod that is used to execute the build
	PodID string `json:"podID,omitempty" yaml:"podID,omitempty"`
}

// BuildInput defines input parameters for a given build
type BuildInput struct {
	//SourceType is the type of SCM used as the build input
	SourceType SourceType `json:"sourceType" yaml:"sourceType"`

	// GitSource is SCM system in use
	GitSource *GitSourceControl `json:"gitSource,omitempty" yaml:"gitSource,omitempty"`

	// ImageTag is the tag to give to the image resulting from the build
	ImageTag string `json:"imageTag,omitempty" yaml:"imageTag,omitempty"`

	// Registry to push the result image to
	Registry string `json:"registry,omitempty" yaml:"registry,omitempty"`

	// DockerBuild represents build parameters specific to docker build
	DockerInput *DockerBuildInput `json:"dockerInput,omitempty" yaml:"dockerInput,omitempty"`

	// STIBuild represents build parameters specific to STI build
	STIInput *STIBuildInput `json:"stiInput,omitempty" yaml:"stiInput,omitempty"`
}

// SourceType is the type of SCM used
type SourceType string

const (
	//GitSource is a Git SCM
	GitSource SourceType = "git"
)

// GitSourceControl defines the parameters of a Git SCM
type GitSourceControl struct {
	// URI points to the source that will be built. The structure of the source
	// will depend on the type of build to run
	URI string `json:"uri,omitempty" yaml:"uri,omitempty"`

	// Ref is the branch/tag/ref to build.
	Ref string `json:"ref,omitempty" yaml:"ref,omitempty"`

	// Commit is the specific commit information to build
	Commit Commit `json:"commit,omitempty" yaml:"commit,omitempty"`
}

// ScmUser defines the identity of a user of source control
type ScmUser struct {
	Name  string `json:"name,omitempty" yaml:"name,omitempty"`
	Email string `json:"email,omitempty" yaml:"email,omitempty"`
}

// Commit is the snapshot information for a specific revision in a ScmRepo
type Commit struct {
	ID        string  `json:"id,omitempty" yaml:"id,omitempty"`
	Author    ScmUser `json:"author,omitempty" yaml:"author,omitempty"`
	Committer ScmUser `json:"committer,omitempty" yaml:"committer,omitempty"`
	Message   string  `json:"message,omitempty" yaml:"message,omitempty"`
}

// DockerBuildInput defines input parameters specific to docker build
type DockerBuildInput struct {
	// ContextDir is a directory inside the SourceURI structure which should be used as a docker
	// context when building
	ContextDir string `json:"contextDir,omitempty" yaml:"contextDir,omitempty"`
}

// STIBuildInput defines input parameters specific to sti build
type STIBuildInput struct {
	// BuilderImage is the image used to execute the build
	BuilderImage string `json:"builderImage,omitempty" yaml:"builderImage,omitempty"`
}

// BuildConfig contains the inputs needed to produce a new deployable image
type BuildConfig struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Labels       map[string]string `json:"labels,omitempty" yaml:"labels,omitempty"`

	// DesiredInput is the input used to create builds from this configuration
	DesiredInput BuildInput `json:"desiredInput,omitempty" yaml:"desiredInput,omitempty"`

	// Secret used to validate requests.
	Secret string `json:"secret,omitempty" yaml:"secret,omitempty"`
}

// BuildType is a type of build (docker, sti, etc)
type BuildType string

// Valid build types
const (
	// DockerBuildType is a build based on a Dockerfile with associated artifacts
	DockerBuildType BuildType = "docker"

	// STIBuildType is a build using Source to Image using a git repository
	// and a builder image
	STIBuildType BuildType = "sti"
)

// BuildStatus represents the status of a Build at a point in time.
type BuildStatus string

// Valid build status values
const (
	// BuildNew is automatically assigned to a newly created build
	BuildNew BuildStatus = "new"

	// BuildPending indicates that a pod name has been assigned and a build is
	// about to start running
	BuildPending BuildStatus = "pending"

	// BuildRunning indicates that a pod has been created and a build is running
	BuildRunning BuildStatus = "running"

	// BuildComplete indicates that a build has been successful
	BuildComplete BuildStatus = "complete"

	// BuildFailed indicates that a build has executed and failed
	BuildFailed BuildStatus = "failed"

	// BuildError indicates that an error prevented the build from
	// executing
	BuildError BuildStatus = "error"
)

// BuildList is a collection of Builds.
type BuildList struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Items        []Build `json:"items,omitempty" yaml:"items,omitempty"`
}

// BuildConfigList is a collection of BuildConfigs.
type BuildConfigList struct {
	api.JSONBase `json:",inline" yaml:",inline"`
	Items        []BuildConfig `json:"items,omitempty" yaml:"items,omitempty"`
}
