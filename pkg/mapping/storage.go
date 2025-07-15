package mapping

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider"
	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// createStorageMapping creates a new storage mapping
func createStorageMapping(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile string) error {
	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	if fromFile != "" {
		return createMappingFromFile(dynamicClient, fromFile, namespace)
	}

	// Create a typed StorageMap
	storageMap := &v1beta1.StorageMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: v1beta1.StorageMapSpec{
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
			Map: []v1beta1.StoragePair{},
		},
	}

	// Convert to unstructured
	unstructuredObj, err := runtime.DefaultUnstructuredConverter.ToUnstructured(storageMap)
	if err != nil {
		return fmt.Errorf("failed to convert to unstructured: %v", err)
	}

	mapping := &unstructured.Unstructured{Object: unstructuredObj}
	mapping.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   client.Group,
		Version: client.Version,
		Kind:    "StorageMap",
	})

	_, err = dynamicClient.Resource(client.StorageMapGVR).Namespace(namespace).Create(context.TODO(), mapping, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create storage mapping: %v", err)
	}

	fmt.Printf("Storage mapping '%s' created in namespace '%s'\n", name, namespace)
	return nil
}
