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

	defaultDCIntervalSec     int64 = 1
	defaultDCTimeoutSec      int64 = 600
	defaultDCUpdatePeriodSec int64 = 1

	defaultElasticHost = "logging-es"
	defaultElasticPort = 9200

	elasticTerminationGracePeriodSec int64 = 600
	elasticMemory                          = "512Mi"

	curatorCPU = "100m"

	defaultMasterURL = "https://kubernetes.default.svc.cluster.local"
)

var componentNames = sets.NewString(componentKibana, componentFluentd, componentCurator, componentElastic)
