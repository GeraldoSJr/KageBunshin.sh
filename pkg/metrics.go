package pkg

import (
    "context"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/metrics/pkg/client/clientset/versioned"
    "k8s.io/api/core/v1"
)

type Metrics struct {
    CpuUsage    resource.Quantity
    MemoryUsage resource.Quantity
}

type PodsMetrics struct {
    PodName string
    Namespace string
    CpuLimit resource.Quantity
    MemoryLimit resource.Quantity
}

// nodeMetrics collects metrics for all nodes in the cluster
func NodeMetrics(clientset *kubernetes.Clientset, metricsClient *versioned.Clientset, ctx context.Context) []Metrics {
    var nodesReturn []Metrics

    // List all nodes in the cluster using the provided context
    nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
    if err != nil {
        panic(err.Error())
    }

    // Loop over the list of nodes and get their metrics
    for _, node := range nodes.Items {
        metrics, err := metricsClient.MetricsV1beta1().NodeMetricses().Get(ctx, node.Name, metav1.GetOptions{})
        if err != nil {
            panic(err.Error())
        }

        // Get CPU and memory usage directly from the metrics API
        cpuUsage := metrics.Usage["cpu"]
        memoryUsage := metrics.Usage["memory"]

        // Append the node's metrics to the result slice
        nodeMetrics := Metrics{
            CpuUsage:    cpuUsage,
            MemoryUsage: memoryUsage,
        }
        nodesReturn = append(nodesReturn, nodeMetrics)
    }

    return nodesReturn
}

func PodMetrics(pods []v1.Pod) ([]PodsMetrics, error) {
    var podLimitsResults []PodsMetrics

    for _, pod := range pods {
        // Check if the pod is in "Pending" state
        if pod.Status.Phase != v1.PodPending {
            continue // Skip non-pending pods
        }

        // Initialize the total CPU and memory limits for the pod
        totalCpuLimit := resource.NewQuantity(0, resource.DecimalSI)
        totalMemoryLimit := resource.NewQuantity(0, resource.BinarySI)

        // Iterate over the containers in the pod spec to get the resource limits
        for _, container := range pod.Spec.Containers {
            cpuLimit, cpuExists := container.Resources.Limits[v1.ResourceCPU]
            memoryLimit, memoryExists := container.Resources.Limits[v1.ResourceMemory]

            // Add CPU limit if it exists
            if cpuExists {
                totalCpuLimit.Add(cpuLimit)
            }

            // Add memory limit if it exists
            if memoryExists {
                totalMemoryLimit.Add(memoryLimit)
            }
        }

        // Append the pod's limit data to the result
        podLimitsResults = append(podLimitsResults, PodsMetrics{
            PodName:    pod.Name,
            Namespace:  pod.Namespace,
            CpuLimit:   *totalCpuLimit,
            MemoryLimit: *totalMemoryLimit,
        })
    }

    return podLimitsResults, nil
}

func PendingPods(clientset *kubernetes.Clientset, ctx context.Context) []v1.Pod {
    var pendingPods []v1.Pod

    pods, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
    if err != nil {
        panic(err.Error())
    }

    for _, pod := range pods.Items {
        if pod.Status.Phase == "Pending" {
            pendingPods = append(pendingPods, pod)
        }
    }

    return pendingPods
}
