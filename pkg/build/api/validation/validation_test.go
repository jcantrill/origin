package validation

import (
	"testing"

	kubeapi "github.com/GoogleCloudPlatform/kubernetes/pkg/api"
	errs "github.com/GoogleCloudPlatform/kubernetes/pkg/api/errors"
	"github.com/openshift/origin/pkg/build/api"
)

func TestBuildValdationSuccess(t *testing.T) {
	build := &api.Build{
		JSONBase: kubeapi.JSONBase{ID: "buildID"},
		Input: api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "http://github.com/my/repository",
			},
			ImageTag: "repository/data",
		},
		Status: api.BuildNew,
	}
	if result := ValidateBuild(build); len(result) > 0 {
		t.Errorf("Unexpected validation error returned %v", result)
	}
}

func TestBuildValidationFailure(t *testing.T) {
	build := &api.Build{
		JSONBase: kubeapi.JSONBase{ID: ""},
		Input: api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "http://github.com/my/repository",
			},
			ImageTag: "repository/data",
		},
		Status: api.BuildNew,
	}
	if result := ValidateBuild(build); len(result) != 1 {
		t.Errorf("Unexpected validation result: %v", result)
	}
}

func TestBuildConfigValidationSuccess(t *testing.T) {
	buildConfig := &api.BuildConfig{
		JSONBase: kubeapi.JSONBase{ID: "configID"},
		DesiredInput: api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "http://github.com/my/repository",
			},
			ImageTag: "repository/data",
		},
	}
	if result := ValidateBuildConfig(buildConfig); len(result) > 0 {
		t.Errorf("Unexpected validation error returned %v", result)
	}
}

func TestBuildConfigValidationFailure(t *testing.T) {
	buildConfig := &api.BuildConfig{
		JSONBase: kubeapi.JSONBase{ID: ""},
		DesiredInput: api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "http://github.com/my/repository",
			},
			ImageTag: "repository/data",
		},
	}
	if result := ValidateBuildConfig(buildConfig); len(result) != 1 {
		t.Errorf("Unexpected validation result %v", result)
	}
}

func TestValidateGitSource(t *testing.T) {
	if result := validateGitSource(nil); len(result) != 1 {
		t.Errorf("Expected validation error when source is missing")
	}
}

func TestValidateBuildInput(t *testing.T) {
	errorCases := map[string]*api.BuildInput{
		string(errs.ValidationErrorTypeRequired) + "gitSource.URI": &api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "",
			},
			ImageTag: "repository/data",
		},
		string(errs.ValidationErrorTypeInvalid) + "gitSource.URI": &api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "::",
			},
			ImageTag: "repository/data",
		},
		string(errs.ValidationErrorTypeRequired) + "imageTag": &api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "http://github.com/test/uri",
			},
			ImageTag: "",
		},
		string(errs.ValidationErrorTypeRequired) + "stiBuild.builderImage": &api.BuildInput{
			GitSource: &api.GitSourceControl{
				URI: "http://github.com/test/uri",
			},
			ImageTag: "repository/data",
			STIInput: &api.STIBuildInput{
				BuilderImage: "",
			},
		},
	}

	for desc, config := range errorCases {
		errors := validateBuildInput(config)
		if len(errors) != 1 {
			t.Errorf("%s: Unexpected validation result: %v", desc, errors)
		}
		err := errors[0].(errs.ValidationError)
		errDesc := string(err.Type) + err.Field
		if desc != errDesc {
			t.Errorf("Unexpected validation result for %s: expected %s, got %s", err.Field, desc, errDesc)
		}
	}
}
