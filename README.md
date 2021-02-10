## k8s-discovery
[![PkgGoDev](https://pkg.go.dev/badge/github.com/gkarthiks/k8s-discovery)](https://pkg.go.dev/github.com/gkarthiks/k8s-discovery)
![Release](https://img.shields.io/github/tag-date/gkarthiks/k8s-discovery.svg?color=Orange&label=Latest%20Release)
![language](https://img.shields.io/badge/Language-go-blue.svg)
[![Go Report Card](https://goreportcard.com/badge/github.com/gkarthiks/k8s-discovery)](https://goreportcard.com/report/github.com/gkarthiks/k8s-discovery)
![License](https://img.shields.io/github/license/gkarthiks/k8s-discovery.svg)

![](go-k8s.png)

*K8s Discovery*, is an effort to reduce the boiler plate code for the [client-go](https://github.com/kubernetes/client-go) or kubernetes based cloud native [GoLang](https://golang.org) developers.

The main aspect of this is around saving and cleaning the code for in-cluster and out-cluster configurations. 


## Usage

Run `go get` to get the *k8s-discovery* module as follows.

```
go get github.com/gkarthiks/k8s-discovery
```

Declare a variable as `var k8s *discovery.K8s` and initialize it as `k8s, _ = discovery.NewK8s()`. Now the **k8s** will hold the interface that will provide the clientset for Kubernetes communication that is pulled either via `in-cluster` or via `kubeconfig` file.

## Available APIs at the moment

<b>NewK8s:</b> Will return a new kubernetes clientset's interface that is formulated either via in-cluster configuration or kubeconfog file.

<b>GetVersion:</b> Queries the Kubernetes for the version in `v0.0.0-master+$Format:%h$`

<b>GetNamespace:</b> Gets the namespace of the running pod if running inside the cluster, if outside returns based on the `POD_NAMESPACE` environment variable. This environment variable also takes precedence if provided in a pod.

## Available client
*K8s Discovery* provides the client set for kubernetes client with hassle free configuration as well as the metrics client. The `MetricsClientSet` can be used to query the metrics against the containers. There is also a `RestConfig` exposed via *discovery* to make use of the rest api.


## Example
```go
package main

import (
	"context"
	"fmt"
	"log"
	"strconv"

	discovery "github.com/gkarthiks/k8s-discovery"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsTypes "k8s.io/metrics/pkg/apis/metrics/v1beta1"
)

var (
	k8s *discovery.K8s
)

func main() {
	k8s, _ = discovery.NewK8s()
	namespace, _ := k8s.GetNamespace()
	version, _ := k8s.GetVersion()
	fmt.Printf("Specified Namespace: %s\n", namespace)
	fmt.Printf("Version of running Kubernetes: %s\n", version)
	
	
	
	
	cronJobs, err := k8s.Clientset.BatchV1beta1().CronJobs(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Panic(err.Error())
	}
	for idx, crons := range cronJobs.Items {
		fmt.Printf("%d -> %s\n", idx, crons.Name)
	}
	
	
	
	
	fmt.Println("==== Moving towards the metrics query ====")
	
	podMetrics, err := k8s.MetricsClientSet.MetricsV1beta1().PodMetricses(namespace).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		panic(err)
	}

	var podMetric metricsTypes.PodMetrics
	getPodUsageMetrics := func(pod metricsTypes.PodMetrics) {
		for _, container := range pod.Containers {
			cpuQuantityDec := container.Usage.Cpu().AsDec().String()
			cpuUsageFloat, _ := strconv.ParseFloat(cpuQuantityDec, 64)

			fmt.Printf( "CPU Usage Float: %v\n", cpuUsageFloat)

			memoryQuantityDec := container.Usage.Memory().AsDec().String()
			memoryUsageFloat, _ := strconv.ParseFloat(memoryQuantityDec, 64)
			fmt.Printf("Memory Usage Float: %v\n\n", memoryUsageFloat)
		}
	}

	for _, podMetric = range podMetrics.Items {
		getPodUsageMetrics(podMetric)
	}
	
	
	
}
```

### Note:

For GCP or managed kubernetes, you have to import the `auth` module, else an error message stating `no Auth Provider found for name "gcp"` will be thrown. The import looks like the below for the sample program. Special mention @ringods.

```golang
import (
	"context"
	"fmt"
	"log"
	"strconv"

	discovery "github.com/gkarthiks/k8s-discovery"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	metricsTypes "k8s.io/metrics/pkg/apis/metrics/v1beta1"
	_ "k8s.io/client-go/plugin/pkg/client/auth"
)
```
