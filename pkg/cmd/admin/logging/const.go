package logging

import (
	"k8s.io/kubernetes/pkg/util/sets"
)

const (
	namePrefixServiceAccount   = "aggregated-logging"
	namePrefixDeploymentConfig = "logging"

	componentFluentd = "fluentd"
	componentCurator = "curator"
	componentElastic = "elasticsearch"
	componentKibana  = "kibana"
)

var componentNames = sets.NewString(componentKibana, componentFluentd, componentCurator, componentElastic)
