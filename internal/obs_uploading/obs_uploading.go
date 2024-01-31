package obs_uploading

import (
	"cloud-terraform-mirror/internal/config"
	"fmt"
	"os"
	"os/exec"
)

func ObsUpload(conf *config.Conf, dirPath string, providerNamespace, providerName string) error {
	conf.Obs.Mu.Lock()
	args := append([]string{"obsutil", "sync"}, fmt.Sprintf("-p=%d", 3), fmt.Sprintf("-j=%d", 1),
		dirPath, fmt.Sprintf("obs://tf-mirror-pub/registry.terraform.io/%s/%s",
			providerNamespace, providerName))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	conf.Obs.Mu.Unlock()
	return nil
}
func ObsUploadingSettings(conf *config.Conf, dirPath string, providerNamespace, providerName string) error {
	conf.Obs.Mu.Lock()
	args := append([]string{"obsutil", "sync"}, fmt.Sprintf("-p=%d", 1), fmt.Sprintf("-j=%d", 1),
		dirPath, fmt.Sprintf("obs://tf-mirror-int/settings/settings_providers/%s-%s.json",
			providerNamespace, providerName))
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err := cmd.Run()
	if err != nil {
		return err
	}

	conf.Obs.Mu.Unlock()
	return nil
}
