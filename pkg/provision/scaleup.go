package provision

import (
    "context"
    "github.com/GeraldoSJr/KageBunshin.sh/pkg"
    "k8s.io/apimachinery/pkg/api/resource"
    "k8s.io/client-go/kubernetes"
    "k8s.io/metrics/pkg/client/clientset/versioned"
)

// NodeMap represents the aggregated CPU and Memory needs of a node
type NodeMap struct {
    CpuNeed    resource.Quantity
    MemoryNeed resource.Quantity
}

// ScaleUp calculates the resources required to schedule pending pods and sends NodeMap objects via a channel
func ScaleUp(ctx context.Context, clientset *kubernetes.Clientset, metricsClient *versioned.Clientset)[]NodeMap {
    var nodeList []NodeMap

    pendingPods := pkg.PendingPods(clientset, ctx)

    podMetrics, err := pkg.PodMetrics(pendingPods)
    if err != nil {
        panic(err.Error())
    }

    var nodeMetrics = NodeMap{
        CpuNeed:    *resource.NewQuantity(0, resource.DecimalSI),
        MemoryNeed: *resource.NewQuantity(0, resource.BinarySI),
    }

    for _, pod := range podMetrics {
        podCpuNeed := pod.CpuLimit
        podMemoryNeed := pod.MemoryLimit

        newCpuTotal := nodeMetrics.CpuNeed.DeepCopy()
        newCpuTotal.Add(podCpuNeed)

        newMemoryTotal := nodeMetrics.MemoryNeed.DeepCopy()
        newMemoryTotal.Add(podMemoryNeed)

        if newCpuTotal.Cmp(resource.MustParse("2")) > 0 || newMemoryTotal.Cmp(resource.MustParse("2Gi")) > 0 {
            nodeList = append(nodeList, nodeMetrics)


            nodeMetrics = NodeMap{
                CpuNeed:    podCpuNeed.DeepCopy(),
                MemoryNeed: podMemoryNeed.DeepCopy(),
            }
        } else {
            nodeMetrics.CpuNeed.Add(podCpuNeed)
            nodeMetrics.MemoryNeed.Add(podMemoryNeed)
        }
    }

    if nodeMetrics.CpuNeed.Cmp(resource.MustParse("0")) > 0 || nodeMetrics.MemoryNeed.Cmp(resource.MustParse("0Gi")) > 0 {
        nodeList = append(nodeList, nodeMetrics)
    }

    return nodeList
}


