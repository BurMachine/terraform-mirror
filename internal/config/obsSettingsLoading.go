package config

import (
	"errors"
	"fmt"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"io"
	"os"
	"path/filepath"
)

// LoadConfig
// Getting config.yaml from obs
func LoadConfig(conf *Conf) error {
	var err error
	conf.Obs.ObsClient, err = obs.New(conf.ObsAccessKey, conf.ObsSecretKey, conf.ObsEndpoint)
	if err != nil {
		fmt.Printf("Create obsClient error, errMsg: %s", err.Error())
	}
	input := &obs.GetObjectInput{}

	input.Bucket = "tf-mirror-int"
	input.Key = "settings/config.yaml"

	output, err := conf.Obs.ObsClient.GetObject(input)
	if err != nil {
		if obsError, ok := err.(obs.ObsError); ok {
			if obsError.StatusCode != 404 {
				return errors.Join(err, errors.New(fmt.Sprintf("input key: %s", input.Key)))
			}
		}
	}
	var body []byte
	if output != nil {
		body, err = io.ReadAll(output.Body)
		if err != nil {
			return err
		}
	} else {
		file, err := os.Open("config-default.yaml")
		if err != nil {
			return err
		}
		defer file.Close()

		body, err = io.ReadAll(file)
		if err != nil {
			return err
		}
	}

	err = os.WriteFile("config.yaml", body, os.ModePerm)
	if err != nil {
		return err
	}
	return nil
}

// LoadSettingsObs
// Load .json settings files if are exist(loading list is config.yaml)
func (c *Conf) LoadSettingsObs() error {
	outputFolder, err := filepath.Abs("output/settings")
	if err != nil {
		return err
	}
	if _, err = os.Stat(outputFolder); os.IsNotExist(err) {
		err = os.MkdirAll(outputFolder, os.ModePerm)
		if err != nil {
			return err
		}
	}

	// Downloading providers settings from obs, if provider not exist - skip him
	for _, provider := range c.GenerateSettingArr.Providers {
		fileName := fmt.Sprintf("%s-%s", provider.Namespace, provider.Name)
		input := &obs.GetObjectInput{}
		input.Bucket = "tf-mirror-int"

		input.Key = fmt.Sprintf("settings/settings_providers/%s.json", fileName)

		output, err := c.Obs.ObsClient.GetObject(input)
		if err != nil {
			if obsError, ok := err.(obs.ObsError); ok {
				if obsError.StatusCode == 404 {
					continue
				} else {
					return errors.Join(err, errors.New(fmt.Sprintf("input key: %s", input.Key)))
				}
			}

		}
		if output != nil {
			body, err := io.ReadAll(output.Body)
			if err != nil {
				return err
			}
			err = os.WriteFile(fmt.Sprintf("%s/%s.json", outputFolder, fileName), body, os.ModePerm)
			if err != nil {
				return err
			}
		} else {
			return errors.New("obs config loading error: empty object output")
		}

	}
	return nil
}
