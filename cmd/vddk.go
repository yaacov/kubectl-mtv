package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/yaacov/kubectl-mtv/pkg/vddk"
)

var (
	vddkTarGz    string
	vddkTag      string
	vddkBuildDir string
	vddkPush     bool
)

func newVddkCmd() *cobra.Command {
	var createCmd = &cobra.Command{
		Use:   "create",
		Short: "Build a VDDK image for MTV",
		RunE: func(cmd *cobra.Command, args []string) error {
			return vddk.BuildImage(vddkTarGz, vddkTag, vddkBuildDir, vddkPush)
		},
	}
	createCmd.Flags().StringVar(&vddkTarGz, "tar", "", "Path to VMware VDDK tar.gz file (required), e.g. VMware-vix-disklib.tar.gz")
	createCmd.Flags().StringVar(&vddkTag, "tag", "", "Container image tag (required), e.g. quay.io/example/vddk:8.0.1")
	createCmd.Flags().StringVar(&vddkBuildDir, "build-dir", "", "Build directory (optional, uses tmp dir if not set)")
	createCmd.Flags().BoolVar(&vddkPush, "push", false, "Push image after build (optional)")

	err := createCmd.MarkFlagRequired("tar")
	if err != nil {
		fmt.Printf("Warning: error marking 'tar' flag as required: %v\n", err)
	}
	err = createCmd.MarkFlagRequired("tag")
	if err != nil {
		fmt.Printf("Warning: error marking 'tag' flag as required: %v\n", err)
	}

	var vddkCmd = &cobra.Command{
		Use:   "vddk",
		Short: "VDDK image utilities",
	}
	vddkCmd.AddCommand(createCmd)
	return vddkCmd
}
