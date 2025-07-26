package generic

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	forkliftv1beta1 "github.com/kubev2v/forklift/pkg/apis/forklift/v1beta1"

	"github.com/yaacov/kubectl-mtv/pkg/commands/provider/providerutil"
	"github.com/yaacov/kubectl-mtv/pkg/util/client"
)

// validateProviderOptions validates the options for creating a generic provider
func validateProviderOptions(options providerutil.ProviderOptions, providerType string) error {
	if options.Name == "" {
		return fmt.Errorf("provider name is required")
	}
	if options.Namespace == "" {
		return fmt.Errorf("provider namespace is required")
	}
	if options.URL == "" {
		return fmt.Errorf("provider URL is required")
	}

	// If token is provided, username and password are not required
	if options.Token == "" {
		if options.Username == "" {
			return fmt.Errorf("provider username is required (unless token is provided)")
		}
		if options.Password == "" {
			return fmt.Errorf("provider password is required (unless token is provided)")
		}
	}

	if options.Secret != "" && (options.Username != "" || options.Password != "" || options.Token != "") {
		return fmt.Errorf("if a secret is provided, username, password, and token should not be specified")
	}
	if options.Secret == "" && options.Token == "" && (options.Username == "" || options.Password == "") {
		return fmt.Errorf("if no secret is provided, either token or username and password must be specified")
	}

	return nil
}

// cleanupCreatedResources deletes any resources created during the provider creation process
func cleanupCreatedResources(configFlags *genericclioptions.ConfigFlags, namespace string, secret *corev1.Secret) {
	if secret != nil {
		c, err := client.GetDynamicClient(configFlags)
		if err != nil {
			return
		}

		err = c.Resource(client.SecretsGVR).Namespace(namespace).Delete(context.TODO(), secret.Name, metav1.DeleteOptions{})
		if err != nil {
			fmt.Printf("Warning: failed to clean up secret %s: %v\n", secret.Name, err)
		}
	}
}

// setSecretOwnership sets the provider as the owner of the secret
func setSecretOwnership(configFlags *genericclioptions.ConfigFlags, provider *forkliftv1beta1.Provider, secret *corev1.Secret) error {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return fmt.Errorf("failed to get client: %v", err)
	}

	// Set the provider as the owner of the secret
	secret.SetOwnerReferences([]metav1.OwnerReference{
		{
			APIVersion: provider.APIVersion,
			Kind:       provider.Kind,
			Name:       provider.Name,
			UID:        provider.UID,
		},
	})

	// Convert secret to unstructured and update
	unstructSecret, err := runtime.DefaultUnstructuredConverter.ToUnstructured(secret)
	if err != nil {
		return fmt.Errorf("failed to convert secret to unstructured: %v", err)
	}

	unstructuredSecret := &unstructured.Unstructured{Object: unstructSecret}
	_, err = c.Resource(client.SecretsGVR).Namespace(secret.Namespace).Update(context.TODO(), unstructuredSecret, metav1.UpdateOptions{})
	if err != nil {
		return fmt.Errorf("failed to update secret ownership: %v", err)
	}

	return nil
}

// createTypedProvider creates an unstructured provider and converts it to a typed Provider
func createTypedProvider(configFlags *genericclioptions.ConfigFlags, namespace string, provider *forkliftv1beta1.Provider) (*forkliftv1beta1.Provider, error) {
	c, err := client.GetDynamicClient(configFlags)
	if err != nil {
		return nil, fmt.Errorf("failed to get client: %v", err)
	}

	// Convert provider to unstructured
	unstructProvider, err := runtime.DefaultUnstructuredConverter.ToUnstructured(provider)
	if err != nil {
		return nil, fmt.Errorf("failed to convert provider to unstructured: %v", err)
	}

	unstructuredProvider := &unstructured.Unstructured{Object: unstructProvider}

	// Create the provider
	createdUnstructProvider, err := c.Resource(client.ProvidersGVR).Namespace(namespace).Create(
		context.TODO(), unstructuredProvider, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to create provider: %v", err)
	}

	// Convert unstructured provider to typed provider
	createdProvider := &forkliftv1beta1.Provider{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(createdUnstructProvider.Object, createdProvider); err != nil {
		return nil, fmt.Errorf("failed to convert provider from unstructured: %v", err)
	}

	return createdProvider, nil
}

// CreateProvider implements a generic provider creator for oVirt and OpenStack
func CreateProvider(configFlags *genericclioptions.ConfigFlags, options providerutil.ProviderOptions, providerType string) (*forkliftv1beta1.Provider, *corev1.Secret, error) {
	// Validate required fields
	if err := validateProviderOptions(options, providerType); err != nil {
		return nil, nil, err
	}

	// Create basic provider structure
	provider := &forkliftv1beta1.Provider{}
	provider.SetName(options.Name)
	provider.SetNamespace(options.Namespace)
	provider.APIVersion = forkliftv1beta1.SchemeGroupVersion.String()
	provider.Kind = "Provider"

	// Set provider type and URL
	providerTypeValue := forkliftv1beta1.ProviderType(providerType)
	provider.Spec.Type = &providerTypeValue
	provider.Spec.URL = options.URL

	// Create and set the Secret
	var createdSecret *corev1.Secret
	var err error

	if options.Secret == "" {
		createdSecret, err = createSecret(configFlags, options.Namespace, options.Name,
			options.Username, options.Password, options.URL, options.CACert, options.Token, options.InsecureSkipTLS,
			options.DomainName, options.ProjectName, options.RegionName, providerType)
		if err != nil {
			return nil, nil, fmt.Errorf("failed to create %s secret: %v", providerType, err)
		}

		provider.Spec.Secret = corev1.ObjectReference{
			Name:      createdSecret.Name,
			Namespace: createdSecret.Namespace,
		}
	} else {
		provider.Spec.Secret = corev1.ObjectReference{
			Name:      options.Secret,
			Namespace: options.Namespace,
		}
	}

	// Create the provider
	createdProvider, err := createTypedProvider(configFlags, options.Namespace, provider)
	if err != nil {
		// Clean up the created secret if provider creation fails
		cleanupCreatedResources(configFlags, options.Namespace, createdSecret)

		return nil, nil, fmt.Errorf("failed to create %s provider: %v", providerType, err)
	}

	// Set the secret ownership to the provider
	if createdSecret != nil {
		if err := setSecretOwnership(configFlags, createdProvider, createdSecret); err != nil {
			return nil, createdSecret, fmt.Errorf("provider created but %v", err)
		}
	}

	return createdProvider, createdSecret, nil
}
