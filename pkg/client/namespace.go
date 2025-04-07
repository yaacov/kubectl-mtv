package client

import (
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// ForkliftMTVNamespace is the default namespace where Forklift MTV CRDs are deployed
const OpenShiftMTVNamespace = "openshift-mtv"

// ResolveNamespace determines the effective namespace with fallback logic:
// 1. Use namespace from command line flags if specified
// 2. Use namespace from current kubeconfig context if available
// 3. Fall back to "default" namespace if neither is available
func ResolveNamespace(configFlags *genericclioptions.ConfigFlags) string {
	// Check if the namespace is set in the command line flags
	if configFlags.Namespace != nil && *configFlags.Namespace != "" {
		return *configFlags.Namespace
	}

	// Try to get the namespace from kubeconfig
	clientConfig := configFlags.ToRawKubeConfigLoader()
	if clientConfig != nil {
		namespace, _, err := clientConfig.Namespace()
		if err == nil && namespace != "" {
			return namespace
		}
	}

	// Fall back to default namespace if not found in kubeconfig
	return "default"
}
