package garbagecollection

import (
	"context"
	"fmt"
	"github.com/hashicorp/go-multierror"

	argov1 "github.com/argoproj/argo/pkg/apis/workflow/v1alpha1"
	"github.com/gardener/test-infra/pkg/testmachinery"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// CleanWorkflowPods deletes all pods of a completed workflow.
// cleanup pods to remove workload from the api server and etcd.
// logs are still accessible through "archiveLogs" option in argo
func CleanWorkflowPods(c client.Client, wf *argov1.Workflow) error {
	var result *multierror.Error
	if testmachinery.GetConfig().CleanWorkflowPods {
		for nodeName, node := range wf.Status.Nodes {
			if node.Type == argov1.NodeTypePod {
				if err := deletePod(c, testmachinery.GetConfig().Namespace, nodeName); err != nil {
					result = multierror.Append(result, fmt.Errorf("unable delete pod %s: %s", nodeName, err.Error()))
				}
			}
		}
	}
	return result.ErrorOrNil()
}

func deletePod(c client.Client, namespace, name string) error {
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}
	return c.Delete(context.TODO(), pod)
}
