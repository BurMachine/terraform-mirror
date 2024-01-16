package config

import (
	"errors"
	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
	"gopkg.in/yaml.v3"
	"io/ioutil"
	"os"
)

type GenerateSettings struct {
	Providers []struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
	} `yaml:"providers"`
}

type Conf struct {
	ObsEndpoint        string           `yaml:"obsEndpoint"`
	ObsAccessKey       string           `yaml:"obsAccessKey"`
	ObsSecretKey       string           `yaml:"obsSecretKey"`
	RegistryAddr       string           `yaml:"RegistryAddr,omitempty"`
	GenerateSettingArr GenerateSettings `yaml:"generate"`

	ObsClient *obs.ObsClient
}

func New() *Conf {
	return &Conf{}
}

var RegistryAddrDefault = "https://registry.terraform.io/v1/providers"

// LoadConfig
// Getting env variables
// Loading config.yaml file from obs
// Loading settings files from obs
func (c *Conf) LoadConfig() error {
	c.ObsEndpoint = os.Getenv("obsEndpoint")
	if c.ObsEndpoint == "" {
		return errors.New("env getting error: obsEndpoint is empty")
	}
	c.ObsAccessKey = os.Getenv("obsAccessKey")
	if c.ObsAccessKey == "" {
		return errors.New("env getting error: obsAccessKey is empty")
	}
	c.ObsSecretKey = os.Getenv("obsSecretKey")
	if c.ObsSecretKey == "" {
		return errors.New("env getting error: obsSecretKey is empty")
	}

	err := LoadConfig(c)
	if err != nil {
		return err
	}

	err = c.LoadGenerateSettingsYaml()
	if err != nil {
		return err
	}
	c.RegistryAddr = os.Getenv("REGISTRY_ADDR")
	if c.RegistryAddr == "" {
		c.RegistryAddr = RegistryAddrDefault
	}

	err = c.LoadSettingsObs()
	if err != nil {
		return err
	}

	return nil
}

func (c *Conf) LoadGenerateSettingsYaml() error {
	file, err := os.Open("config.yaml")
	if err != nil {
		return err
	}
	defer file.Close()

	data, err := ioutil.ReadAll(file)
	if err != nil {
		return err
	}

	var settings GenerateSettings
	if err := yaml.Unmarshal(data, &settings); err != nil {
		return err
	}

	c.GenerateSettingArr = settings
	return nil
}
