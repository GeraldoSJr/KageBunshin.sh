package provision

import (
    "context"
    "k8s.io/client-go/kubernetes"
    pkg "github.com/GeraldoSJr/KageBunshin.sh/pkg"
)

func ScaleDown(clientset *kubernetes.Clientset, ctx context.Context) error {
    nodeMetrics := pkg.NodeMetrics(clientset, ctx)

    for _, node := range nodeMetrics {
        if node.CpuUsage.Value() == 0 && node.MemoryUsage.Value() == 0 {
            // Scale down the node
            }
        }
    return nil
}

