package main

import (
    "context"
    "fmt"
    "flag"
    "path/filepath"

    "github.com/GeraldoSJr/KageBunshin.sh/pkg/provision"
    "k8s.io/apimachinery/pkg/api/resource"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
)

const cpuThreshold = "800m"
const memoryThreshold = "2Gi"

type NodeMap struct {
    CpuNeed resource.Quantity
    MemoryNeed resource.Quantity
}

func main() {
	clientset := createK8SClient()
    ctx := context.TODO()

    nodeList := provision.ScaleUp(ctx, clientset)

    fmt.Println("==== Cluster Node Resource Metrics ====")
    for i, newNodeMetrics := range nodeList {
        fmt.Printf("Node %d:\n", i+1)
        fmt.Printf("  - CPU Need: %s\n", newNodeMetrics.CpuNeed.String())
        fmt.Printf("    Explanation: This node requires %s CPUs to accommodate pending pods.\n", newNodeMetrics.CpuNeed.String())
        fmt.Printf("  - Memory Need: %s\n", newNodeMetrics.MemoryNeed.String())
        fmt.Printf("    Explanation: This node requires %s of memory to accommodate pending pods.\n", newNodeMetrics.MemoryNeed.String())
    }
    fmt.Println("==== End of Metrics ====")

}
func resourceMustParse(value string) *resource.Quantity {
    quantity := resource.MustParse(value)
    return &quantity
}

func createK8SClient() *kubernetes.Clientset {
    var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
    config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
    if err != nil {
        panic(err.Error())
    }

    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        panic(err.Error())
    }

    return clientset
}


