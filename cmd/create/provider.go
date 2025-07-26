package create

import (
	"os"
	"strings"

	"github.com/spf13/cobra"
	"k8s.io/cli-runtime/pkg/genericclioptions"

	"github.com/yaacov/kubectl-mtv/pkg/client"
	"github.com/yaacov/kubectl-mtv/pkg/flags"
	"github.com/yaacov/kubectl-mtv/pkg/provider"
)

// NewProviderCmd creates the provider creation command
func NewProviderCmd(kubeConfigFlags *genericclioptions.ConfigFlags) *cobra.Command {
	var secret string
	providerType := flags.NewProviderTypeFlag()

	// Add Provider credential flags
	var url, username, password, cacert, token string
	var insecureSkipTLS bool
	var vddkInitImage, sdkEndpoint string

	// OpenStack specific flags
	var domainName, projectName, regionName string

	// Check if MTV_VDDK_INIT_IMAGE environment variable is set
	if envVddkInitImage := os.Getenv("MTV_VDDK_INIT_IMAGE"); envVddkInitImage != "" {
		vddkInitImage = envVddkInitImage
	}

	cmd := &cobra.Command{
		Use:          "provider NAME",
		Short:        "Create a new provider",
		Args:         cobra.ExactArgs(1),
		SilenceUsage: true,
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
				url, username, password, cacert, insecureSkipTLS, vddkInitImage, sdkEndpoint, token,
				domainName, projectName, regionName)
		},
	}

	cmd.Flags().Var(providerType, "type", "Provider type (openshift, vsphere, ovirt, openstack, ova)")
	cmd.Flags().StringVar(&secret, "secret", "", "Secret containing provider credentials")

	// Provider credential flags
	cmd.Flags().StringVarP(&url, "url", "U", "", "Provider URL")
	cmd.Flags().StringVarP(&username, "username", "u", "", "Provider credentials username")
	cmd.Flags().StringVarP(&password, "password", "p", "", "Provider credentials password")
	cmd.Flags().StringVar(&cacert, "cacert", "", "Provider CA certificate (use @filename to load from file)")
	cmd.Flags().BoolVar(&insecureSkipTLS, "provider-insecure-skip-tls", false, "Skip TLS verification when connecting to the provider")

	// OpenShift specific flags
	cmd.Flags().StringVarP(&token, "token", "T", "", "Provider authentication token (used for openshift provider)")

	// VSphere specific flags
	cmd.Flags().StringVar(&vddkInitImage, "vddk-init-image", vddkInitImage, "Virtual Disk Development Kit (VDDK) container init image path")
	cmd.Flags().StringVar(&sdkEndpoint, "sdk-endpoint", "", "SDK endpoint type for vSphere provider (vcenter or esxi)")

	// OpenStack specific flags
	cmd.Flags().StringVar(&domainName, "provider-domain-name", "", "OpenStack domain name")
	cmd.Flags().StringVar(&projectName, "provider-project-name", "", "OpenStack project name")
	cmd.Flags().StringVar(&regionName, "provider-region-name", "", "OpenStack region name")

	// Add completion for provider type flag
	if err := cmd.RegisterFlagCompletionFunc("type", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return providerType.GetValidValues(), cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	// Add completion for sdk-endpoint flag
	if err := cmd.RegisterFlagCompletionFunc("sdk-endpoint", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		return []string{"vcenter", "esxi"}, cobra.ShellCompDirectiveNoFileComp
	}); err != nil {
		panic(err)
	}

	if err := cmd.MarkFlagRequired("type"); err != nil {
		panic(err)
	}

	return cmd
}
