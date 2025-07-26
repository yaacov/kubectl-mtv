package mapping

import (
	"context"
	"fmt"
	"strings"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"
	"github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1/provider"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
	corev1 "k8s.io/api/core/v1"
)

// parseNetworkPairs parses network pairs in format "source1:namespace/target1,source2:namespace/target2"
// If namespace is omitted, the provided defaultNamespace will be used
// Special target values: "pod" for pod networking, "ignored" to ignore the source network
func parseNetworkPairs(pairStr, defaultNamespace string, configFlags *genericclioptions.ConfigFlags, sourceProvider, inventoryURL string) ([]v1beta1.NetworkPair, error) {
	if pairStr == "" {
		return nil, nil
	}

	var pairs []v1beta1.NetworkPair
	pairList := strings.Split(pairStr, ",")

	for _, pairStr := range pairList {
		pairStr = strings.TrimSpace(pairStr)
		if pairStr == "" {
			continue
		}

		parts := strings.SplitN(pairStr, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid network pair format '%s': expected 'source:target-namespace/target-network', 'source:target-network', 'source:pod', or 'source:ignored'", pairStr)
		}

		sourceName := strings.TrimSpace(parts[0])
		targetPart := strings.TrimSpace(parts[1])

		// Resolve source network name to ID
		sourceNetworkRef, err := resolveNetworkNameToID(configFlags, sourceProvider, defaultNamespace, inventoryURL, sourceName)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve source network '%s': %v", sourceName, err)
		}

		// Parse target part which can be just a name or namespace/name
		var targetNamespace, targetName, targetType string
		if strings.Contains(targetPart, "/") {
			targetParts := strings.SplitN(targetPart, "/", 2)
			targetNamespace = strings.TrimSpace(targetParts[0])
			targetName = strings.TrimSpace(targetParts[1])
			targetType = "multus"
		} else {
			// Special handling for 'pod' and 'ignored' types
			switch targetPart {
			case "pod":
				targetType = "pod"
			case "ignored":
				targetType = "ignored"
			default:
				// Use the target part as network name and default namespace
				targetName = targetPart
				targetNamespace = defaultNamespace
				targetType = "multus"
			}
		}

		destinationNetwork := v1beta1.DestinationNetwork{
			Type: targetType,
		}
		if targetName != "" {
			destinationNetwork.Name = targetName
		}
		if targetNamespace != "" {
			destinationNetwork.Namespace = targetNamespace
		}

		pair := v1beta1.NetworkPair{
			Source:      sourceNetworkRef,
			Destination: destinationNetwork,
		}

		pairs = append(pairs, pair)
	}

	return pairs, nil
}

// createNetworkMapping creates a new network mapping
func createNetworkMapping(configFlags *genericclioptions.ConfigFlags, name, namespace, sourceProvider, targetProvider, fromFile, networkPairs, inventoryURL string) error {
	dynamicClient, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	if fromFile != "" {
		return createMappingFromFile(dynamicClient, fromFile, namespace)
	}

	// Parse network pairs if provided
	var mappingPairs []v1beta1.NetworkPair
	if networkPairs != "" {
		mappingPairs, err = parseNetworkPairs(networkPairs, namespace, configFlags, sourceProvider, inventoryURL)
		if err != nil {
			return fmt.Errorf("failed to parse network pairs: %v", err)
		}
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
			Map: mappingPairs,
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
