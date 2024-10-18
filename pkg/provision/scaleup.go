package provision

import (
    "context"
    "github.com/GeraldoSJr/KageBunshin.sh/pkg"
    "k8s.io/apimachinery/pkg/api/resource"
    "k8s.io/client-go/kubernetes"
)

// NodeMap represents the aggregated CPU and Memory needs of a node
type NodeMap struct {
    CpuNeed    resource.Quantity
    MemoryNeed resource.Quantity
}

func ScaleUp(ctx context.Context, clientset *kubernetes.Clientset) []NodeMap {
    var nodeList []NodeMap

    // Get all pending pods
    pendingPods := pkg.PendingPods(clientset, ctx)

    // Retrieve the pod metrics (e.g., CPU and memory limits)
    podMetrics, err := pkg.PodMetrics(pendingPods)
    if err != nil {
        panic(err.Error())
    }

    // Initialize a NodeMap to track CPU and memory usage for a new node
    var nodeMetrics = NodeMap{
        CpuNeed:    *resource.NewQuantity(0, resource.DecimalSI),
        MemoryNeed: *resource.NewQuantity(0, resource.BinarySI),
    }

    for _, pod := range podMetrics {
        podCpuNeed := pod.CpuLimit
        podMemoryNeed := pod.MemoryLimit

        // Calculate the total CPU and memory needs if this pod is added to the current node
        newCpuTotal := nodeMetrics.CpuNeed.DeepCopy()
        newCpuTotal.Add(podCpuNeed)

        newMemoryTotal := nodeMetrics.MemoryNeed.DeepCopy()
        newMemoryTotal.Add(podMemoryNeed)

        // If the new totals exceed node limits, save the current node and start a new one
        if newCpuTotal.Cmp(resource.MustParse("2")) > 0 || newMemoryTotal.Cmp(resource.MustParse("2Gi")) > 0 {
            if nodeMetrics.CpuNeed.Cmp(resource.MustParse("0")) > 0 || nodeMetrics.MemoryNeed.Cmp(resource.MustParse("0")) > 0 {
                nodeList = append(nodeList, nodeMetrics) // Add only if there is a non-zero resource need
            }

            // Start a new node for the next set of resources
            nodeMetrics = NodeMap{
                CpuNeed:    podCpuNeed.DeepCopy(),
                MemoryNeed: podMemoryNeed.DeepCopy(),
            }
        } else {
            // Otherwise, add the pod's resource needs to the current node
            nodeMetrics.CpuNeed.Add(podCpuNeed)
            nodeMetrics.MemoryNeed.Add(podMemoryNeed)
        }
    }

    // Add the last node if it has non-zero resource needs
    if nodeMetrics.CpuNeed.Cmp(resource.MustParse("0")) > 0 || nodeMetrics.MemoryNeed.Cmp(resource.MustParse("0")) > 0 {
        nodeList = append(nodeList, nodeMetrics)
    }

    return nodeList
}
