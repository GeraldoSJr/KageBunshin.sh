package main

import (
    "context"
    "fmt"
    "log"
    "path/filepath"

    pkg "github.com/GeraldoSJr/KageBunshin.sh/pkg" // Assuming nodeMetrics function is in "pkg" package
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/clientcmd"
    "k8s.io/client-go/util/homedir"
    "k8s.io/metrics/pkg/client/clientset/versioned"
)

const cpuThreshold = "800m"
const memoryThreshold = "2Gi"

func main() {
    // Step 1: Create Kubernetes and Metrics clients
    clientset, metricsClient := createK8sClients()

    // Step 2: Create a context to pass to the nodeMetrics function
    ctx := context.TODO()

    // Step 3: Get node metrics using the imported nodeMetrics function
    nodeMetricsList := pkg.NodeMetrics(clientset, metricsClient, ctx)

    // Step 4: Display node metrics
    fmt.Println("==== Cluster Node Resource Metrics ====")
    for _, nodeMetrics := range nodeMetricsList {
        cpuMillicores := nodeMetrics.CpuUsage.MilliValue()
        memoryMiB := nodeMetrics.MemoryUsage.ScaledValue(resource.Mega)

        fmt.Printf("\nNode Metrics:\n")
        fmt.Printf("  - CPU Usage: %d m\n", cpuMillicores)
        fmt.Printf("    Explanation: The current CPU usage for this node is %d millicores (m).\n", cpuMillicores)
        fmt.Printf("  - Memory Usage: %d MiB\n", memoryMiB)
        fmt.Printf("    Explanation: The current memory usage for this node is %d MiB (Mebibytes).\n", memoryMiB)

        // Check thresholds and take action
        if nodeMetrics.CpuUsage.Cmp(*resourceMustParse(cpuThreshold)) > 0 {
            fmt.Printf("\n[ALERT] CPU usage is above threshold (%s).\n", cpuThreshold)
            fmt.Println("Suggested Action: Consider scaling up the node resources.")
        }
        if nodeMetrics.MemoryUsage.Cmp(*resourceMustParse(memoryThreshold)) > 0 {
            fmt.Printf("\n[ALERT] Memory usage is above threshold (%s).\n", memoryThreshold)
            fmt.Println("Suggested Action: Consider scaling up the node resources.")
        }
    }

    // Step 5: Get pod list and check for any pending pods
    pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        log.Fatalf("Error retrieving pods: %v", err)
    }
    for _, pod := range pods.Items {
        if pod.Status.Phase == "Pending" {
            fmt.Printf("\nPod %s is in Pending state. Scaling up the node.\n", pod.Name)
        }
    }

    fmt.Println("==== End of Metrics ====")
}

func resourceMustParse(value string) *resource.Quantity {
    quantity := resource.MustParse(value)
    return &quantity
}

func createK8sClients() (*kubernetes.Clientset, *versioned.Clientset) {
    var config *rest.Config
    var err error

    // Check if we are running inside a Kubernetes cluster or not
    if home := homedir.HomeDir(); home != "" {
        kubeconfig := filepath.Join(home, ".kube", "config")
        config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
        if err != nil {
            log.Fatalf("Error building kubeconfig: %v", err)
        }
    } else {
        config, err = rest.InClusterConfig()
        if err != nil {
            log.Fatalf("Error creating in-cluster config: %v", err)
        }
    }

    // Create Kubernetes clientset
    clientset, err := kubernetes.NewForConfig(config)
    if err != nil {
        log.Fatalf("Error creating Kubernetes clientset: %v", err)
    }

    // Create Metrics clientset
    metricsClient, err := versioned.NewForConfig(config)
    if err != nil {
        log.Fatalf("Error creating Metrics clientset: %v", err)
    }

    return clientset, metricsClient
}

