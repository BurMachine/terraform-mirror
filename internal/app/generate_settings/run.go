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
		go func(name, namespace, registryUrl string) {
			defer wg.Done()
			logger.Logger.Infof("generate version JSON files for namespace: %s, provider: %s", namespace, name)
			err := serviceProc(name, namespace, registryUrl)
			if err != nil {
				errChan <- err
			}
		}(provider.Name, provider.Namespace, conf.RegistryAddr)
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

func serviceProc(name, namespace, registryUrl string) error {
	url := fmt.Sprintf("%s/%s/%s/versions", registryUrl, namespace, name)
	resp, err := http.Get(url)
	if err != nil || resp.StatusCode != 200 {
		if resp.StatusCode != 200 {
			return fmt.Errorf("request [%s], status: %s", url, resp.Status)
		}
		return err
	}
	defer resp.Body.Close()

	outputFolder := "./output/settings"

	if _, err := os.Stat(outputFolder); os.IsNotExist(err) {
		err := os.MkdirAll(outputFolder, os.ModePerm)
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

	_, err = io.Copy(outputFile, resp.Body)
	if err != nil {
		return err
	}

	return nil
}
