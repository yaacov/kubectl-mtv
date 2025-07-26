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
	"sigs.k8s.io/yaml"
)

// CreateMappingFromFile creates a mapping from a YAML file
func CreateMappingFromFile(dynamicClient dynamic.Interface, fileName, namespace string) error {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return fmt.Errorf("failed to read file: %v", err)
	}

	obj := &unstructured.Unstructured{}
	if err := json.Unmarshal(data, &obj.Object); err != nil {
		// Try YAML if JSON fails
		if yamlErr := yaml.Unmarshal(data, &obj.Object); yamlErr != nil {
			return fmt.Errorf("failed to parse file as JSON or YAML: JSON error: %v, YAML error: %v", err, yamlErr)
		}
	}

	// Set namespace if not specified in the file
	if obj.GetNamespace() == "" && namespace != "" {
		obj.SetNamespace(namespace)
	}

	// Determine the appropriate GVR based on the object kind
	var gvr schema.GroupVersionResource
	switch obj.GetKind() {
	case "NetworkMap":
		gvr = schema.GroupVersionResource{
			Group:    "forklift.konveyor.io",
			Version:  "v1beta1",
			Resource: "networkmaps",
		}
	case "StorageMap":
		gvr = schema.GroupVersionResource{
			Group:    "forklift.konveyor.io",
			Version:  "v1beta1",
			Resource: "storagemaps",
		}
	default:
		return fmt.Errorf("unsupported kind: %s", obj.GetKind())
	}
	_, err = dynamicClient.Resource(gvr).Namespace(obj.GetNamespace()).Create(context.TODO(), obj, metav1.CreateOptions{})
	if err != nil {
		return fmt.Errorf("failed to create mapping: %v", err)
	}

	fmt.Printf("Mapping '%s' created in namespace '%s'\n", obj.GetName(), obj.GetNamespace())
	return nil
}
