package mapping

import (
	"context"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1"
	"github.com/konveyor/forklift-controller/pkg/apis/forklift/v1beta1/provider"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	corev1 "k8s.io/api/core/v1"
)

// createNetworkMapping creates a new network mapping
func createNetworkMapping(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile string) error {
	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	if fromFile != "" {
		return createMappingFromFile(dynamicClient, fromFile, namespace)
	}

	// Create a typed NetworkMap
	networkMap := &v1beta1.NetworkMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.NetworkMapSpec{
			Provider: provider.Pair{
				Source: corev1.ObjectReference{
					Name:      sourceProvider,
					Namespace: namespace,
				},
				Destination: corev1.ObjectReference{
					Name:      targetProvider,
					Namespace: namespace,
				},
			},
			Map: []v1beta1.NetworkPair{},
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(networkMap)
	if err != nil {
		return fmt.Errorf("failed to convert to unstructured: %v", err)
	}

	mapping := &unstructured.Unstructured{Object: unstructuredObj}
	mapping.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   client.Group,
		Version: client.Version,
		Kind:    "NetworkMap",
	})

	_, err = dynamicClient.Resource(client.NetworkMapGVR).Namespace(namespace).Create(context.TODO(), mapping, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create network mapping: %v", err)
	}

	fmt.Printf("Network mapping '%s' created in namespace '%s'\n", name, namespace)
	return nil
}
