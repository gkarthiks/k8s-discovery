package discovery

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
	coreClient "k8s.io/client-go/kubernetes"
	restClient "k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	cmdClient "k8s.io/client-go/tools/clientcmd"
	metricsClient "k8s.io/metrics/pkg/client/clientset/versioned"
)

// K8s struct holds the instance of clientset and metrics clisentset
type K8s struct {
	Clientset        coreClient.Interface
	MetricsClientSet *metricsClient.Clientset
	RestConfig       *restClient.Config
	inKluster        bool
	isDevEnv         bool
	currentContext   string // only for out-of-kluster config
}

// dev/production env detection, copied from Tilt
type Env string

const (
	EnvUnknown       Env = "unknown"
	EnvGKE           Env = "gke"
	EnvMinikube      Env = "minikube"
	EnvDockerDesktop Env = "docker-for-desktop"
	EnvMicroK8s      Env = "microk8s"
	EnvCRC           Env = "crc"
	EnvKrucible      Env = "krucible"
	// Kind v0.6 substantially changed the protocol for detecting and pulling,
	// so we represent them as two separate envs.
	EnvKIND5          Env = "kind-0.5-"
	EnvKIND6          Env = "kind-0.6+"
	EnvK3D            Env = "k3d"
	EnvRancherDesktop Env = "rancher-desktop"
	EnvNone           Env = "none" // k8s not running (not neces. a problem, e.g. if using Tilt x Docker Compose)
)

var logEnabled bool

// NewK8s will provide a new k8s client interface
// resolves where it is running whether inside the kubernetes cluster or outside
// While running outside of the cluster, tries to make use of the kubeconfig file
// While running inside the cluster resolved via pod environment uses the in-cluster config
func NewK8s() (*K8s, error) {
	client := K8s{}

	_, logEnabled = os.LookupEnv("CLIENTSET_LOG")

	client.inKluster = true
	client.isDevEnv = false
	config, err := restClient.InClusterConfig()
	if err != nil {
		client.inKluster = false
		kubeConfig :=
			cmdClient.NewDefaultClientConfigLoadingRules().GetDefaultFilename()
		config, err = cmdClient.BuildConfigFromFlags("", kubeConfig)
		if err != nil {
			return nil, err
		}

		err = client.addContextInfo(kubeConfig)
		if err != nil {
			return nil, err
		}

		if logEnabled {
			log.Info("Program running from outside of the cluster")
		}
	} else {
		if logEnabled {
			log.Info("Program running inside the cluster, picking the in-cluster configuration")
		}
	}
	client.Clientset, err = coreClient.NewForConfig(config)
	client.MetricsClientSet, err = metricsClient.NewForConfig(config)
	client.RestConfig = config

	return &client, nil
}

// GetVersion returns the version of the kubernetes cluster that is running
func (o *K8s) GetVersion() (string, error) {
	version, err := o.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	if logEnabled {
		log.Infof("Version of running k8s %v", version)
	}
	return fmt.Sprintf("%s", version), nil
}

// GetNamespace will return the current namespace for the running program
// Checks for the user passed ENV variable POD_NAMESPACE if not available
// pulls the namespace from pod, if not returns ""
func (o *K8s) GetNamespace() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		}
		return "", err
	}
	return "", nil
}

func (o *K8s) IsInKluster() bool {
	return o.inKluster
}

func (o *K8s) IsDevEnv() bool {
	return o.isDevEnv
}

func (o *K8s) CurrentContext() string {
	return o.currentContext
}

func (e Env) isDevCluster() bool {
	return e == EnvMinikube ||
		e == EnvDockerDesktop ||
		e == EnvMicroK8s ||
		e == EnvCRC ||
		e == EnvKIND5 ||
		e == EnvKIND6 ||
		e == EnvK3D ||
		e == EnvKrucible ||
		e == EnvRancherDesktop
}

func (o *K8s) addContextInfo(kubeConfig string) (err error) {
	// write down current context
	rawconfig, err := clientcmd.NewNonInteractiveDeferredLoadingClientConfig(
		&clientcmd.ClientConfigLoadingRules{ExplicitPath: kubeConfig},
		&clientcmd.ConfigOverrides{
			CurrentContext: "",
		}).RawConfig()
	if err != nil {
		return
	}
	o.currentContext = rawconfig.CurrentContext
	env := Env(o.currentContext)
	o.isDevEnv = env.isDevCluster()
	return
}
