package createMirror

import (
	"cloud-terraform-mirror/internal/models"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"slices"
	"strings"
)

const mirrorFolder = "output/mirror"

func processing(module *models.Module, logger *loggerLogrus.Logger, exitChan chan struct{}) error {
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
					//err = os.Remove(fmt.Sprintf("tmp_%s.json", n[1]))
					//if err != nil {
					//	if err.Error() != fmt.Sprintf("remove tmp_%s.json: no such file or directory", n[1]) {
					//		exitChan <- struct{}{}
					//		return err
					//	}
					//}
					exitChan <- struct{}{}
					return nil
				default:
					continue
				}
			}

		}
	}
	// check hash-sums
	/*

	 */

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
		//err := os.Remove(fmt.Sprintf("tmp_%s.json", n[1]))
		//if err != nil {
		//	if err.Error() != fmt.Sprintf("remove tmp_%s.json: no such file or directory", n[1]) {
		//		exitChan <- struct{}{}
		//		return err
		//	}
		//}
		exitChan <- struct{}{}
		return nil
	default:
	}
	//err := os.Remove(fmt.Sprintf("tmp_%s.json", n[1]))
	//if err != nil {
	//	if err.Error() != fmt.Sprintf("remove tmp_%s.json: no such file or directory", n[1]) {
	//		return err
	//	}
	//}
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
