package generic

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
)

// createSecret creates a secret for generic providers (oVirt, OpenStack)
func createSecret(configFlags *genericclioptions.ConfigFlags, namespace, providerName, user, password, url, cacert string, insecureSkipTLS bool, domainName, projectName, regionName, providerType string) (*corev1.Secret, error) {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %v", err)
	}

	secretName := fmt.Sprintf("%s-provider-secret", providerName)

	// Prepare secret data
	secretData := map[string][]byte{
		"user":     []byte(user),
		"password": []byte(password),
		"url":      []byte(url),
	}

	// Add CA certificate if provided
	if cacert != "" {
		secretData["cacert"] = []byte(cacert)
	}

	// Add insecureSkipVerify if true
	if insecureSkipTLS {
		secretData["insecureSkipVerify"] = []byte("true")
	}

	// Add OpenStack specific fields if provided
	if providerType == "openstack" {
		if domainName != "" {
			secretData["domainName"] = []byte(domainName)
		}
		if projectName != "" {
			secretData["projectName"] = []byte(projectName)
		}
		if regionName != "" {
			secretData["regionName"] = []byte(regionName)
		}
	}

	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      secretName,
			Namespace: namespace,
		},
		Type: corev1.SecretTypeOpaque,
		Data: secretData,
	}

	// Convert secret to unstructured
	unstructSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return nil, fmt.Errorf("failed to convert secret to unstructured: %v", err)
	}

	unstructuredSecret := &unstructured.Unstructured{Object: unstructSecret}

	// Create the secret
	createdUnstructSecret, err := c.Resource(client.SecretsGVR).Namespace(namespace).Create(context.TODO(), unstructuredSecret, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create secret: %v", err)
	}

	// Convert unstructured secret back to typed secret
	createdSecret := &corev1.Secret{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(createdUnstructSecret.Object, createdSecret); err != nil {
		return nil, fmt.Errorf("failed to convert secret from unstructured: %v", err)
	}

	return createdSecret, nil
}
