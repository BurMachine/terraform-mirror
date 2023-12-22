package createMirror

import (
	"cloud-terraform-mirror/internal/models"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const mirrorFolder = "output/mirror"

func processing(module *models.Module, logger *loggerLogrus.Logger) error {
	if _, err := os.Stat(mirrorFolder); os.IsNotExist(err) {
		err = os.MkdirAll(mirrorFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	n := strings.Split(module.ID, "/")

	for _, version := range module.Versions {
		if !slices.Contains(version.Protocols, "4") && !slices.Contains(version.Protocols, "4.0") {
			for _, p := range version.Platforms {
				platform := fmt.Sprintf("%s_%s", p.OS, p.Arch)
				err := createMainTF(n[1], n[0], version.Version)
				if err != nil {
					return err
				}
				err = terraformMirror(platform)
				if err != nil {
					return err
				}
			}
		}

	}
	return nil
}

func createMainTF(providerName, providerNamespace, version string) error {
	content := fmt.Sprintf(`terraform {
  required_providers {
    %s = {
      source  = "%s/%s"
      version = "%s"
    }
  }
}`, providerName, providerNamespace, providerName, version)

	return os.WriteFile("main.tf", []byte(content), os.ModePerm)
}

func terraformMirror(platform string) error {
	cmd := exec.Command("terraform", "providers", "mirror", "-platform="+platform, "./output/mirror")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
