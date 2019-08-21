package discovery

import (
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	log "github.com/sirupsen/logrus"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type K8s struct {
	Clientset kubernetes.Interface
}

// NewK8s will provide a new k8s client interface
// resolves where it is running whether inside the kubernetes cluster or outside
// While running outside of the cluster, tries to make use of the kubeconfig file
// While running inside the cluster resolved via pod environment uses the in-cluster config
func NewK8s() (*K8s, error) {
	client := K8s{}
	if _, inCluster := os.LookupEnv("KUBERNETES_SERVICE_HOST"); inCluster == true {
		log.Info("Program running inside the cluster, picking the in-cluster configuration")

		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		client.Clientset, err = kubernetes.NewForConfig(config)
		if err != nil {
			return nil, err
		}
		return &client, nil
	}

	log.Info("Program running from outside of the cluster")
	var kubeconfig *string
	if home := homeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		return nil, err
	}
	client.Clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	return &client, nil
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE")
}

// GetVersion returns the version of the kubernetes cluster that is running
func (o *K8s) GetVersion() (string, error) {
	version, err := o.Clientset.Discovery().ServerVersion()
	if err != nil {
		return "", err
	}
	log.Println(version)
	return fmt.Sprintf("%s", version), nil
}

// GetNamespace will return the current namespace for the running program
// Checks for the user passed ENV variable POD_NAMESPACE if not available
// pulls the namespace from pod, if not returns ""
func GetNamespace() (string, error) {
	if ns := os.Getenv("POD_NAMESPACE"); ns != "" {
		return ns, nil
	}
	if data, err := ioutil.ReadFile("/var/run/secrets/kubernetes.io/serviceaccount/namespace"); err == nil {
		if ns := strings.TrimSpace(string(data)); len(ns) > 0 {
			return ns, nil
		} else {
			return "", err
		}
	}
	return "", nil
}