package generateSettings

import (
	"cloud-terraform-mirror/internal/config"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"sync"
)

func Run(conf *config.Conf, logger *loggerLogrus.Logger) error {
	var wg sync.WaitGroup
	errChan := make(chan error, len(conf.GenerateSettingArr.Providers))

	for _, provider := range conf.GenerateSettingArr.Providers {
		wg.Add(1)
		go func(name, namespace, minVersion string, registryUrl string) {
			defer wg.Done()
			logger.Logger.Infof("generate version JSON files for namespace: %s, provider: %s", namespace, name)
			err := serviceProc(name, namespace, minVersion, registryUrl)
			if err != nil {
				errChan <- err
				return
			}
			logger.Logger.Infof("version file for namespace: %s, provider: %s generated successfylly", namespace, name)

		}(provider.Name, provider.Namespace, provider.MinVersion, conf.RegistryAddr)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func serviceProc(name, namespace, minVersion string, registryUrl string) error {
	url := fmt.Sprintf("%s/%s/%s/versions", registryUrl, namespace, name)
	resp, err := http.Get(url)
	if err != nil {
		return err
	} else if resp.StatusCode != 200 {
		return fmt.Errorf("request [%s], status: %s", url, resp.Status)
	}

	outputFolder, err := filepath.Abs("output/settings")
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	err, settingsFileEqual := checkExistVersions(bodyBytes, outputFolder, namespace, minVersion, name)
	if err != nil {
		return err
	} else if settingsFileEqual {
		return nil
	}

	if _, err = os.Stat(outputFolder); os.IsNotExist(err) {
		err = os.MkdirAll(outputFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	outputFilePath := filepath.Join(outputFolder, fmt.Sprintf("%s-%s.json", namespace, name))
	outputFile, err := os.Create(outputFilePath)
	if err != nil {
		return err
	}
	defer outputFile.Close()

	err = os.WriteFile(outputFilePath, bodyBytes, 0644)
	if err != nil {
		return err
	}

	return nil
}
