package createMirror

import (
	"cloud-terraform-mirror/internal/config"
	"cloud-terraform-mirror/internal/models"
	"cloud-terraform-mirror/internal/obs_uploading"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"slices"
	"sort"
	"strings"
	"time"
)

const mirrorFolder = "output/mirror"

func processing(conf *config.Conf, module *models.Module, logger *loggerLogrus.Logger, exitChan chan struct{}) error {

	logger.Logger.Infof("starting creating %s", module.ID)

	if _, err := os.Stat(mirrorFolder); os.IsNotExist(err) {
		err = os.MkdirAll(mirrorFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	start := time.Now()
	duration := time.Since(start)

	defer logger.Logger.Infof("%s processed for %s", module.ID, duration)

	n := strings.Split(module.ID, "/")

	for _, version := range module.Versions {
		logger.Logger.Infof("loading version %s:%s", module.ID, version)
		if !slices.Contains(version.Protocols, "4") && !slices.Contains(version.Protocols, "4.0") {
			for _, p := range version.Platforms {
				platform := fmt.Sprintf("%s_%s", p.OS, p.Arch)

				conf.Obs.Mu.Lock()
				err := createMainTF(n[1], n[0], version.Version)
				if err != nil {
					return err
				}
				conf.Obs.Mu.Unlock()

				err = terraformMirror(platform)
				if err != nil {
					if strings.Contains(err.Error(), "https://registry.terraform.io/.well-known/terraform.json") {
						continue
					}
					if _, err := os.Stat(fmt.Sprintf("tmp_%s.json", n[1])); err != nil {
						if os.IsNotExist(err) {
							if _, err = os.Create(fmt.Sprintf("tmp_%s.json", n[1])); err != nil {
								return err
							}
						}
						return err
					}
					data, err := os.ReadFile(fmt.Sprintf("tmp_%s.json", n[1]))
					if err != nil {
						return err
					}
					errVersions := []models.ErrorVersions{}

					if len(data) != 0 {
						err = json.Unmarshal(data, &errVersions)
						if err != nil {
							return err
						}
					}

					errVersions = append(errVersions, models.ErrorVersions{
						Provider: n[1],
						Version:  version.Version,
						Platform: platform,
					})

					resData, err := json.Marshal(errVersions)
					if err != nil {
						return err
					}

					err = os.WriteFile(fmt.Sprintf("tmp_%s.json", n[1]), resData, os.ModePerm)
					if err != nil {
						return err
					}

				}

				select {
				case <-exitChan:
					exitChan <- struct{}{}
					return nil
				default:
					continue
				}
			}

		}
	}

	// second download attempt
	logger.Logger.Info("start downloading previously undownloaded files...")
	if _, err := os.Stat(fmt.Sprintf("tmp_%s.json", n[1])); !os.IsNotExist(err) {
		data, err := os.ReadFile(fmt.Sprintf("tmp_%s.json", n[1]))
		if err != nil {
			return err
		}
		errVersions := []models.ErrorVersions{}

		if len(data) != 0 {
			if err = json.Unmarshal(data, &errVersions); err != nil {
				return err
			}
			for i, version := range errVersions {
				if err = createMainTF(n[1], n[0], version.Version); err != nil {
					logger.Logger.Error(fmt.Sprintf("%s/%s error second downoad: %v", n[0], n[1], err))
					return err
				}
				if err = terraformMirror(version.Platform); err != nil {
					logger.Logger.Error(fmt.Sprintf("%s/%s error second downoad: %v", n[0], n[1], err))
					return err
				}
				logger.Logger.Info(fmt.Sprintf("%s/%s provider downloaded successfully", n[0], n[1]))
				errVersions[i] = models.ErrorVersions{Provider: "ok"}
			}
		}
	}
	select {
	case <-exitChan:
		exitChan <- struct{}{}
		return nil
	default:
	}

	// obs uploading
	logger.Logger.Info("Starting OBS uploading")
	dirPath := fmt.Sprintf("output/mirror/registry.terraform.io/%s/%s/", n[0], n[1])

	err := loadNewIndex(n[1], n[0])
	if err != nil {
		return err
	}

	err = obs_uploading.ObsUpload(conf, dirPath, n[0], n[1])
	if err != nil {
		err = obs_uploading.ObsUpload(conf, dirPath, n[0], n[1])
		if err != nil {
			return err
		}
	}
	err = obs_uploading.ObsUploadingSettings(conf, fmt.Sprintf("output/settings/%s-%s.json", n[0], n[1]), n[0], n[1])
	if err != nil {
		err = obs_uploading.ObsUploadingSettings(conf, fmt.Sprintf("output/settings/%s-%s.json", n[0], n[1]), n[0], n[1])
		if err != nil {
			return err
		}
	}

	err = loadNewIndex(n[1], n[0])
	if err != nil {
		return err
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
	args := append([]string{"terraform", "providers", "mirror"}, fmt.Sprintf("-platform=%s", platform))
	args = append(args, "./output/mirror")

	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

func loadNewIndex(name, namespace string) error {
	resp, err := http.Get(fmt.Sprintf("https://registry.terraform.io/v1/providers/%s/%s/versions", namespace, name))
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to fetch data: HTTP %d", resp.StatusCode)
	}

	var data map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		return err
	}

	var versionNumbers []string
	for _, v := range data["versions"].([]interface{}) {
		version := v.(map[string]interface{})
		protocols := version["protocols"].([]interface{})
		if len(protocols) == 0 || !containsProtocol(protocols, "4.0") {
			versionNumbers = append(versionNumbers, version["version"].(string))
		}
	}

	sort.Strings(versionNumbers)

	// Создаем словарь с отсортированными версиями
	versions := make(map[string]interface{})
	for _, v := range versionNumbers {
		versions[v] = make(map[string]interface{})
	}
	indexData := map[string]interface{}{"versions": versions}

	jsonData, err := json.MarshalIndent(indexData, "", "    ")
	if err != nil {
		return err
	}

	err = os.WriteFile(fmt.Sprintf("output/mirror/registry.terraform.io/%s/%s/index.json", namespace, name), jsonData, 0644)
	if err != nil {
		return err
	}

	return nil
}

func containsProtocol(protocols []interface{}, target string) bool {
	for _, p := range protocols {
		if p == target {
			return true
		}
	}
	return false
}
