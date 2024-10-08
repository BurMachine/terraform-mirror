package clean

import (
	"os"
)

// Clean
// Cleaning output and config file before start
func Clean(logFileName string) error {
	if _, err := os.Stat("output/"); err == nil {
		if err := os.RemoveAll("output/"); err != nil {
			return err
		}
	}

	if _, err := os.Stat("config.yaml"); err == nil {
		if err := os.Remove("config.yaml"); err != nil {
			return err
		}
	}
	if _, err := os.Stat("main.tf"); err == nil {
		if err := os.Remove("main.tf"); err != nil {
			return err
		}
	}

	if _, err := os.Stat(logFileName); err == nil {
		if err := os.Remove(logFileName); err != nil {
			return err
		}
	}

	return nil
}
