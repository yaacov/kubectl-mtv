package vddk

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// BuildImage builds (and optionally pushes) a VDDK image for MTV.
func BuildImage(tarGzPath, tag, buildDir string, push bool) error {
	if buildDir == "" {
		tmp, err := os.MkdirTemp("", "vddk-build-*")
		if err != nil {
			return fmt.Errorf("failed to create temp dir: %w", err)
		}
		defer os.RemoveAll(tmp)
		buildDir = tmp
	}
	fmt.Printf("Using build directory: %s\n", buildDir)

	// Unpack tar.gz
	fmt.Println("Extracting VDDK tar.gz...")
	if err := extractTarGz(tarGzPath, buildDir); err != nil {
		return fmt.Errorf("failed to extract tar.gz: %w", err)
	}

	// Find the extracted directory
	var distribDir string
	files, _ := os.ReadDir(buildDir)
	for _, f := range files {
		if strings.HasPrefix(f.Name(), "vmware-vix-disklib-distrib") && f.IsDir() {
			distribDir = f.Name()
			break
		}
	}
	if distribDir == "" {
		return fmt.Errorf("could not find vmware-vix-disklib-distrib directory after extraction")
	}

	// Write Dockerfile
	dockerfile := filepath.Join(buildDir, "Dockerfile")
	df := `FROM registry.access.redhat.com/ubi8/ubi-minimal
USER 1001
COPY vmware-vix-disklib-distrib /vmware-vix-disklib-distrib
RUN mkdir -p /opt
ENTRYPOINT ["cp", "-r", "/vmware-vix-disklib-distrib", "/opt"]
`
	if err := os.WriteFile(dockerfile, []byte(df), 0644); err != nil {
		return fmt.Errorf("failed to write Dockerfile: %w", err)
	}

	// Build image
	fmt.Println("Building image with podman...")
	buildCmd := exec.Command("podman", "build", ".", "-t", tag)
	buildCmd.Dir = buildDir
	buildCmd.Stdout = os.Stdout
	buildCmd.Stderr = os.Stderr
	if err := buildCmd.Run(); err != nil {
		return fmt.Errorf("podman build failed: %w", err)
	}

	// Optionally push
	if push {
		fmt.Println("Pushing image...")
		pushCmd := exec.Command("podman", "push", tag)
		pushCmd.Stdout = os.Stdout
		pushCmd.Stderr = os.Stderr
		if err := pushCmd.Run(); err != nil {
			return fmt.Errorf("podman push failed: %w", err)
		}
	}

	fmt.Println("VDDK image build complete.")
	return nil
}

func extractTarGz(tarGzPath, destDir string) error {
	f, err := os.Open(tarGzPath)
	if err != nil {
		return err
	}
	defer f.Close()
	gzr, err := gzip.NewReader(f)
	if err != nil {
		return err
	}
	defer gzr.Close()
	tr := tar.NewReader(gzr)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		target := filepath.Join(destDir, hdr.Name)
		switch hdr.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			outFile, err := os.Create(target)
			if err != nil {
				return err
			}
			if _, err := io.Copy(outFile, tr); err != nil {
				outFile.Close()
				return err
			}
			outFile.Close()
		case tar.TypeSymlink:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			if err := os.Symlink(hdr.Linkname, target); err != nil {
				return err
			}
		}
	}
	return nil
}
