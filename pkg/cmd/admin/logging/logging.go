package logging

import (
	"fmt"
	"io"
	"os"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	kapi "k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	kcmdutil "k8s.io/kubernetes/pkg/kubectl/cmd/util"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/runtime"

	cmdutil "github.com/openshift/origin/pkg/cmd/util"
	"github.com/openshift/origin/pkg/cmd/util/clientcmd"
	configcmd "github.com/openshift/origin/pkg/config/cmd"
	deployapi "github.com/openshift/origin/pkg/deploy/api"
)

const loggingExample = `
`

const loggingLongDesc = `
Install the Aggregated Logging (EFK) components

This command installs EFK components to support the cluster by creating the
required resources to support Aggregated Logging.
`

// Config contains the configuration parameters
// to deploy the aggregated logging stack
type Config struct {
	Action configcmd.BulkAction

	// command config
	DryRun           bool
	ImagesPrefix     string
	UseLatestImages  bool
	ImagesPullSecret string
	Namespace        string

	CACrt string
	CAKey string

	PublicMasterURL string
	ServerTLSJSON   string
	StorageGroup    uint

	// component config
	CuratorConfig
	FluentdConfig
	ElasticSearchConfig
	KibanaConfig

	// internal helpers
	factory *clientcmd.Factory
	out     io.Writer
	cmd     *cobra.Command
}

// CuratorConfig is the config necessary to create curator.
type CuratorConfig struct {
	CuratorNodeSelector    []string
	CuratorOpsNodeSelector []string
}

// Bind binds the flags for curator configuration.
func (cfg *CuratorConfig) Bind(flag *pflag.FlagSet) {
	flag.StringSliceVar(&cfg.CuratorNodeSelector, "curator-nodeselector", []string{}, "A node selector that specifies which nodes are eligible targets for deploying curator instances.")
	flag.StringSliceVar(&cfg.CuratorOpsNodeSelector, "curator-ops-nodeselector", []string{}, "A node selector that specifies which nodes are eligible targets for deploying curator instances to support operations.")
}

// FluentdConfig is the config necessary to create fluentd.
type FluentdConfig struct {
	FluentdOpsNodeSelector []string
}

// Bind binds the flags for fluentd configuration.
func (cfg *FluentdConfig) Bind(flag *pflag.FlagSet) {
	flag.StringSliceVar(&cfg.FluentdOpsNodeSelector, "fluentd-ops-nodeselector", []string{}, "TODO")
}

// ElasticSearchConfig is the config necessary to create elasticsearch.
type ElasticSearchConfig struct {
	ESClusterSize     uint
	ESNodeSelector    []string
	ESPvcDynamic      bool
	ESPvcPrefix       string
	ESPvcSize         string
	ESUseLocalStorage bool

	ESOpsClusterSize  uint
	ESOpsNodeSelector []string
	ESOpsPvcPrefix    string
	ESOpsPvcSize      string
}

// Bind binds the flags for elasticsearch configuration.
func (cfg *ElasticSearchConfig) Bind(flag *pflag.FlagSet) {
	flag.UintVar(&cfg.ESClusterSize, "es-cluster-size", 1, "The number of Elasticsearch nodes to deploy. A minimum of 3 is required for redundacy.")
	flag.StringSliceVar(&cfg.ESNodeSelector, "es-nodeselector", []string{}, "A node selector that specifies which nodes are eligible targets for deploying Elasticsearch instances.")
	flag.BoolVar(&cfg.ESPvcDynamic, "es-pvc-dynamic", false, "Set to true to dynamically provision the pvc backing storage (if available to the cluster)")
	flag.StringVar(&cfg.ESPvcPrefix, "es-pvc-prefix", "logging-es", "Prefix for the names of the pvc for backing storage. A number will be appended per Elasticsearch instance (e.g. logging-es-1).  The pvc will be created if it does not exist with the size defined by --es-pvc-size.")
	flag.StringVar(&cfg.ESPvcSize, "es-pvc-size", "", "The size of the pvc to create per Elasticsearch instance (e.g. 100G).  No pvc will be created if ommited and ephemeral volumes are used instead.")
	flag.BoolVar(&cfg.ESUseLocalStorage, "es-use-local-storage", false, "Prepare Elasticsearch with the appropriate permissions to use local storage (i.e direct volume or pvc).")

	flag.UintVar(&cfg.ESOpsClusterSize, "es-ops-cluster-size", 1, "The number of Elasticsearch nodes to deploy to support operations. A minimum of 3 is required for redundacy.")
	flag.StringSliceVar(&cfg.ESOpsNodeSelector, "es-ops-nodeselector", []string{}, "A node selector that specifies which nodes are eligible targets for deploying Elasticsearch instances to support operations.")
	flag.StringVar(&cfg.ESOpsPvcPrefix, "es-ops-pvc-prefix", "logging-ops-es", "Prefix for the names of the pvc for backing storage to support operations. A number will be appended per Elasticsearch instance (e.g. logging-es-1).  The pvc will be created if it does not exist with the size defined by --es-pvc-size.")
	flag.StringVar(&cfg.ESOpsPvcSize, "e-ops-pvc-size", "", "The size of the pvc to create per Elasticsearch instance (e.g. 100G) to support operations.  No pvc will be created if ommited and ephemeral volumes are used instead.")
}

// KibanaConfig is the config necessary to create kibana.
type KibanaConfig struct {
	KibanaCert         string
	KibanaKey          string
	KibanaNodeSelector []string
	KibanaHostname     string

	KibanaOpsHostname     string
	KibanaOpsCert         string
	KibanaOpsKey          string
	KibanaOpsNodeSelector []string
}

// Bind binds the flags for kibana configuration.
func (cfg *KibanaConfig) Bind(flag *pflag.FlagSet) {
	flag.StringVar(&cfg.KibanaHostname, "kibana-hostname", "kibana.example.com", "The external host name for web clients to reach Kibana.")
	flag.StringVar(&cfg.KibanaCert, "kibana-crt", "", "The filename to a browser facing certificate to the Kibana user interface. Generated if not provided. Default is to generate.")
	flag.StringVar(&cfg.KibanaKey, "kibana-key", "", "The filename to a key to be used with the Kibana certificate. Default is to generate.")
	flag.StringSliceVar(&cfg.KibanaNodeSelector, "kibana-nodeselector", []string{}, "A node selector that specifies which nodes are eligible targets for deploying Kibana instances.")

	flag.StringVar(&cfg.KibanaOpsHostname, "kibana-ops-hostname", "kibana.ops.example.com", "The external host name for web clients to reach Kibana to support operations.")
	flag.StringVar(&cfg.KibanaOpsCert, "kibana-ops-crt", "", "The filename to a browser facing certificate to the Kibana user interface to support operations. Default is to generate.")
	flag.StringVar(&cfg.KibanaOpsKey, "kibana-ops-key", "", "The filename to a key to be used with the Kibana ops certificate. Default is to generate.")
	flag.StringSliceVar(&cfg.KibanaOpsNodeSelector, "kibana-ops-nodeselector", []string{}, "A node selector that specifies which nodes are eligible targets for deploying Kibana instances to support operations.")
}

// NewCmdLogging implements the OpenShift CLI admin logging command.
func NewCmdLogging(f *clientcmd.Factory, parentName, name string, out io.Writer) *cobra.Command {
	cfg := &Config{}
	cmd := &cobra.Command{
		Use:     fmt.Sprintf("%s", name),
		Short:   "Install Aggregated Logging Stack",
		Long:    loggingLongDesc,
		Example: fmt.Sprintf(loggingExample),
		Run: func(cmd *cobra.Command, args []string) {
			kcmdutil.CheckErr(cfg.Complete(f, out, cmd))
			kcmdutil.CheckErr(cfg.Validate())

			err := cfg.RunCmdLogging()
			if err != cmdutil.ErrExit {
				kcmdutil.CheckErr(err)
			} else {
				os.Exit(1)
			}
		},
	}

	flag := cmd.Flags()
	cfg.CuratorConfig.Bind(flag)
	cfg.ElasticSearchConfig.Bind(flag)
	cfg.FluentdConfig.Bind(flag)
	cfg.KibanaConfig.Bind(flag)

	// cmd.Flags().BoolVar(&cfg.DryRun, "dry-run", false, "The the result of the operation without executing it.")
	flag.StringVar(&cfg.ImagesPrefix, "images", "openshift/origin-", "The image prefix to use for retrieving the images. This will be used or all components (e.g. openshift/origin-fluentd:v1.3)")
	flag.StringVar(&cfg.ImagesPullSecret, "images-pull-secret", "", "The name of an existing secret to be used for pulling component images from an authenticated registry.")
	flag.StringVar(&cfg.CACrt, "ca-crt", "", "The filename to a certificate for a CA that will be used to sign any generated certificates. Default is to generate.")
	flag.StringVar(&cfg.CAKey, "ca-key", "", "The filename to a key that matches the certificate specified by --ca-crt. Default is to generate.")
	flag.StringVar(&cfg.PublicMasterURL, "public-master-url", "", "The external URL for the master used to perform OAuth authorizations (e.g. accessing Kibana).  Defaults to the value in kubeconfig.")
	flag.StringVar(&cfg.ServerTLSJSON, "server-tls-json", "", "The filename to a JSON file specifying Node.js TLS options to override the Kibana proxy server defaults.")
	flag.UintVar(&cfg.StorageGroup, "storage-group", 65534, "The number of a supplemental group ID for access to Elasticsearch storage volumes; backing volumes should allow access by this group ID.")

	cfg.Action.BindForOutput(flag)
	return cmd
}

// Complete fills in Config needed if the command is actually invoked.
func (cfg *Config) Complete(f *clientcmd.Factory, out io.Writer, cmd *cobra.Command) error {
	cfg.factory = f
	cfg.out = out
	cfg.cmd = cmd

	return nil
}

// Validate ensures all necessary info is available before running.
func (o *Config) Validate() error {
	return nil
}

// RunCmdLogging is the entry point for the aggregated logging command.
func (cfg *Config) RunCmdLogging() error {
	objects := []runtime.Object{}
	objects = append(objects, createServiceAccounts()...)
	objects = append(objects, createKibana(cfg)...)
	objects = append(objects, createCurator(cfg)...)
	objects = append(objects, createElasticSearch(cfg)...)
	objects = append(objects, createFluentd(cfg)...)

	list := &kapi.List{Items: objects}

	if cfg.Action.ShouldPrint() {
		mapper, _ := cfg.factory.Object(false)
		fn := cmdutil.VersionedPrintObject(cfg.factory.PrintObject, cfg.cmd, mapper, cfg.out)
		if err := fn(list); err != nil {
			return fmt.Errorf("unable to print object: %v", err)
		}
		return nil //defaultOutputErr
	}

	if errs := cfg.Action.WithMessage("Creating logging ...", "created").Run(list, cfg.Namespace); len(errs) > 0 {
		return cmdutil.ErrExit
	}
	return nil
}

// createServiceAccounts creates service accounts for logging components.
func createServiceAccounts() []runtime.Object {
	objects := make([]runtime.Object, componentNames.Len())
	for i, component := range componentNames.List() {
		sa := &kapi.ServiceAccount{
			ObjectMeta: kapi.ObjectMeta{
				Name: fmt.Sprintf("%s-%s", namePrefixServiceAccount, component),
			},
		}
		objects[i] = sa
	}
	return objects
}

// createKibana creates kibana components.
func createKibana(cfg *Config) []runtime.Object {
	deployName := componentKibana
	// TODO version comes from where?
	image := fmt.Sprintf("%slogging-kibana:%s", cfg.ImagesPrefix, "version")
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
							Name:  componentKibana,
							Image: image,
						},
					},
				},
			},
		},
	}
	return []runtime.Object{dc}
}

// createFluentd creates fluentd components.
func createFluentd(cfg *Config) []runtime.Object {
	return nil
}

// createElasticSearch creates elasticsearch components.
func createElasticSearch(cfg *Config) []runtime.Object {
	labels := labels.Set(map[string]string{
		"provider":   "openshift",
		"component":  componentElastic,
		"deployment": "", //TODO where does this come from?
	})

	dc := &deployapi.DeploymentConfig{
		ObjectMeta: kapi.ObjectMeta{
			Name:   fmt.Sprintf("%s-%s", namePrefixDeploymentConfig, componentElastic),
			Labels: labels,
		},
		Spec: deployapi.DeploymentConfigSpec{
			Strategy: deployapi.DeploymentStrategy{
				Type: deployapi.DeploymentStrategyTypeRecreate,
			},
			Template: &kapi.PodTemplateSpec{
				ObjectMeta: kapi.ObjectMeta{Name: componentElastic, Labels: labels},
				Spec: kapi.PodSpec{
					TerminationGracePeriodSeconds: int64Ptr(elasticTerminationGracePeriodSec),
					ServiceAccountName:            fmt.Sprintf("%s-%s", namePrefixServiceAccount, componentFluentd),
					SecurityContext: &kapi.PodSecurityContext{
						SupplementalGroups: []int64{}, // TODO where do these come from
					},
					Containers: []kapi.Container{
						{
							Name: componentElastic,
							// TODO version
							Image: fmt.Sprintf("%slogging-elasticsearch:%s", cfg.ImagesPrefix, "version"),
							Resources: kapi.ResourceRequirements{
								Limits: map[kapi.ResourceName]resource.Quantity{
									// TODO
									kapi.ResourceMemory: resource.MustParse(""),
								},
								Requests: map[kapi.ResourceName]resource.Quantity{
									kapi.ResourceMemory: resource.MustParse(elasticMemory),
								},
							},
							Ports: []kapi.ContainerPort{
								{Name: "restapi", ContainerPort: 9200},
								{Name: "cluster", ContainerPort: 9300},
							},
							Env: []kapi.EnvVar{
								{Name: "KUBERNETES_TRUST_CERT", Value: "true"},
								{Name: "SERVICE_DNS", Value: fmt.Sprintf("logging-%s-cluster", componentElastic)},
								{Name: "CLUSTER_NAME", Value: fmt.Sprintf("logging-%s", componentElastic)},
							},
							// TODO volumeMounts
						},
						// TODO volumes
					},
				},
			},
		},
	}

	return []runtime.Object{dc}
}

// createCurator creates curator components.
func createCurator(cfg *Config) []runtime.Object {
	return nil
}

// int64Ptr is a helper to get the pointer of i.
func int64Ptr(i int64) *int64 {
	return &i
}
