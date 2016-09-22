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

	defaultDCIntervalSec int64 = 1
	defaultDCTimeoutSec int64 = 600
	defaultDCUpdatePeriodSec int64 = 1

	elasticTerminationGracePeriodSec int64 = 600
	elasticMemory = "512Mi"
)

var componentNames = sets.NewString(componentKibana, componentFluentd, componentCurator, componentElastic)
