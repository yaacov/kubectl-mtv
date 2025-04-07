package mapping

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// createMappingFromFile creates a mapping from a YAML file
func createMappingFromFile(dynamicClient dynamic.Interface, fileName, namespace string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(data, &obj.Object); err != nil {
		return fmt.Errorf("failed to parse YAML: %v", err)
	}

	// Set namespace if not specified in the file
	if obj.GetNamespace() == "" {
		obj.SetNamespace(namespace)
	}

	// get object kind
	kind, _, err := unstructured.NestedString(obj.Object, "kind")
	if err != nil {
		return fmt.Errorf("failed to get kind: %v", err)
	}
	if kind != "NetworkMap" && kind != "StorageMap" {
		return fmt.Errorf("invalid kind: %s, expected NetworkMap or StorageMap", kind)
	}

	// Set the GVR based on the kind
	var gvr schema.GroupVersionResource

	if kind == "NetworkMap" {
		gvr = client.NetworkMapGVR
	} else if kind == "StorageMap" {
		gvr = client.StorageMapGVR
	}
	_, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create mapping: %v", err)
	}

	fmt.Printf("Mapping '%s' created in namespace '%s'\n", obj.GetName(), obj.GetNamespace())
	return nil
}

// extractProviderName gets a provider name from the mapping spec
func extractProviderName(mapping unstructured.Unstructured, providerType string) string {
	provider, found, _ := unstructured.NestedMap(mapping.Object, "spec", "provider", providerType)
	if !found || provider == nil {
		return ""
	}

	if name, ok := provider["name"].(string); ok {
		return name
	}
	return ""
}

// createMappingItem creates a standardized mapping item for output
func createMappingItem(mapping unstructured.Unstructured, mappingType string) map[string]interface{} {
	item := map[string]interface{}{
		"name":      mapping.GetName(),
		"namespace": mapping.GetNamespace(),
		"type":      mappingType,
		"source":    extractProviderName(mapping, "source"),
		"target":    extractProviderName(mapping, "destination"),
		"created":   mapping.GetCreationTimestamp().Format("2006-01-02 15:04:05"),
	}

	// Add owner information if available
	if len(mapping.GetOwnerReferences()) > 0 {
		ownerRef := mapping.GetOwnerReferences()[0]
		item["owner"] = ownerRef.Name
		item["ownerKind"] = ownerRef.Kind
	}

	return item
}
