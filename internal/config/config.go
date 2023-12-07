package config

import (
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
	RegistryAddr       string           `yaml:"RegistryAddr,omitempty"`
	GenerateSettingArr GenerateSettings `yaml:"generate"`
}

func New() *Conf {
	return &Conf{}
}

var RegistryAddrDefault = "https://registry.terraform.io/v1/providers"

func (c *Conf) LoadConfig() error {
	err := c.LoadGenerateSettingsYaml()
	if err != nil {
		return err
	}
	c.RegistryAddr = os.Getenv("REGISTRY_ADDR")
	if c.RegistryAddr == "" {
		c.RegistryAddr = RegistryAddrDefault
	}
	//if  {
	//	confString, err := json.Marshal(*c)
	//	if err != nil {
	//		return err
	//	}
	//	return errors.New(fmt.Sprintf("failed get config from environment, there are empty variables: %s", confString))
	//}
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
