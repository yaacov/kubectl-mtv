package create

import (
	"fmt"

	"github.com/spf13/cobra"

	"github.com/yaacov/kubectl-mtv/pkg/cmd/create/vddk"
)

// NewVddkCmd creates the VDDK image creation command
func NewVddkCmd() *cobra.Command {
	var vddkTarGz, vddkTag, vddkBuildDir string
	var vddkPush bool

	cmd := &cobra.Command{
		Use:   "vddk-image",
		Short: "Create a VDDK image for MTV",
		RunE: func(cmd *cobra.Command, args []string) error {
			err := vddk.BuildImage(vddkTarGz, vddkTag, vddkBuildDir, vddkPush)
			if err != nil {
				fmt.Printf("Error building VDDK image: %v\n", err)
				fmt.Printf("You can use the '--help' flag for more information on usage.\n")
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&vddkTarGz, "tar", "", "Path to VMware VDDK tar.gz file (required), e.g. VMware-vix-disklib.tar.gz")
	cmd.Flags().StringVar(&vddkTag, "tag", "", "Container image tag (required), e.g. quay.io/example/vddk:8.0.1")
	cmd.Flags().StringVar(&vddkBuildDir, "build-dir", "", "Build directory (optional, uses tmp dir if not set)")
	cmd.Flags().BoolVar(&vddkPush, "push", false, "Push image after build (optional)")

	if err := cmd.MarkFlagRequired("tar"); err != nil {
		panic(err)
	}
	if err := cmd.MarkFlagRequired("tag"); err != nil {
		panic(err)
	}

	return cmd
}
