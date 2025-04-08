package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/flags"
	"github.com/yaacov/kubectl-mtv/pkg/provider"
)

func newProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "provider",
		Short: "Manage virtualization providers",
		Long:  `Manage providers like oVirt, VMware, OpenStack, and KubeVirt targets`,
	}

	cmd.AddCommand(newCreateProviderCmd())
	cmd.AddCommand(newListProviderCmd())
	cmd.AddCommand(newDeleteProviderCmd())

	return cmd
}

func newCreateProviderCmd() *cobra.Command {
	var secret string
	providerType := flags.NewProviderTypeFlag()

	// Add Provider credential flags
	var url, username, password, cacert, token string
	var insecureSkipTLS bool
	var vddkInitImage string

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a new provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			// Check if cacert starts with @ and load from file if so
			if strings.HasPrefix(cacert, "@") {
				filePath := cacert[1:]
				fileContent, err := os.ReadFile(filePath)
				if err != nil {
					return err
				}
				cacert = string(fileContent)
			}

			return provider.Create(kubeConfigFlags, providerType.GetValue(), name, namespace, secret,
				url, username, password, cacert, insecureSkipTLS, vddkInitImage, token)
		},
	}

	cmd.Flags().Var(providerType, "type", "Provider type (openshift, vsphere, ovirt, openstack, ova)")
	cmd.Flags().StringVar(&secret, "secret", "", "Secret containing provider credentials")

	// Provider credential flags
	cmd.Flags().StringVarP(&url, "url", "U", "", "Provider URL")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Provider credentials username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Provider credentials password")
	cmd.Flags().StringVarP(&token, "token", "T", "", "Provider authentication token (used for openshift provider)")
	cmd.Flags().StringVar(&cacert, "cacert", "", "Provider CA certificate (use @filename to load from file)")
	cmd.Flags().BoolVar(&insecureSkipTLS, "provider-insecure-skip-tls", false, "Skip TLS verification when connecting to the provider")
	cmd.Flags().StringVar(&vddkInitImage, "vddk-init-image", "", "Virtual Disk Development Kit (VDDK) container init image path")

	if err := cmd.MarkFlagRequired("type"); err != nil {
		fmt.Printf("Warning: error marking 'type' flag as required: %v\n", err)
	}

	return cmd
}

func newListProviderCmd() *cobra.Command {
	var inventoryBaseURL string
	var outputFormat string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List providers",
		RunE: func(cmd *cobra.Command, args []string) error {
			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			return provider.List(kubeConfigFlags, namespace, inventoryBaseURL, outputFormat)
		},
	}

	cmd.Flags().StringVarP(&inventoryBaseURL, "inventory-url", "i", "", "Base URL for the inventory service")
	cmd.Flags().StringVarP(&outputFormat, "output", "o", "table", "Output format. One of: table, json")

	return cmd
}

func newDeleteProviderCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			// Get name from positional argument
			name := args[0]

			// Resolve the appropriate namespace based on context and flags
			namespace := client.ResolveNamespace(kubeConfigFlags)

			return provider.Delete(kubeConfigFlags, name, namespace)
		},
	}

	return cmd
}
