package config

import (
	"os"
)

type Conf struct {
	RegistryAddr string
}

func New() *Conf {
	return &Conf{}
}

var RegistryAddrDefault = "https://registry.terraform.io/v1/providers"

func (c *Conf) LoadConfig() error {
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
