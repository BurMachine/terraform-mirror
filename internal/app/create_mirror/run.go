package createMirror

import (
	"cloud-terraform-mirror/internal/config"
	"cloud-terraform-mirror/internal/models"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const deltaDir = "output/settings/deltas"

func Run(conf *config.Conf, logger *loggerLogrus.Logger, exitChan chan struct{}) error {
	logger.Logger.Info("starting creating local mirror")

	fileCount := 0

	dataChan := make(chan models.Module)
	errChan := make(chan error, 100) // 100 for example, depends on count of files
	doneChan := make(chan struct{})
	defer close(dataChan)
	//defer close(errChan)
	defer close(doneChan)

	var wg sync.WaitGroup

	go func() {
		for data := range dataChan {
			go func(data models.Module) {
				wg.Add(1)
				defer wg.Done()
				if len(data.Versions) != 0 {
					err := processing(conf, &data, logger, exitChan)
					if err != nil {
						errChan <- err
						return
					}
					n := strings.Split(data.ID, "/")
					os.Remove(fmt.Sprintf("tmp_%s.json", n[1]))
					fmt.Println("rm", n[1])
				}

				doneChan <- struct{}{}
			}(data)
		}
	}()

	wg.Add(1)
	go func() {
		i := 0
		e := 0
		for {
			select {
			case <-doneChan:
				i++
				if i == fileCount {
					wg.Done()
					return
				}

			case err := <-errChan:
				if err != nil {
					errChan <- err
					e++
					if i+e == fileCount {
						wg.Done()
					}
					return
				}
			}
		}
	}()

	err := filepath.WalkDir(deltaDir, func(fp string, fi os.DirEntry, err error) error {
		if err != nil {
			errChan <- fmt.Errorf("error walking directory: %w", err)
			return nil
		}
		if fi.IsDir() {
			return nil
		}

		wg.Add(1)

		go visitFile(fp, dataChan, errChan, &wg)

		fileCount++
		return nil
	})

	if err != nil {
		return fmt.Errorf("error walking directory: %w", err)
	}

	wg.Wait()
	close(errChan)
	var resErr error
	for err := range errChan {
		resErr = errors.Join(err, resErr)
	}
	return resErr
}

func visitFile(fp string, dataChan chan models.Module, errChan chan error, wg *sync.WaitGroup) {
	defer wg.Done()

	res := models.Module{}

	content, err := os.ReadFile(fp)
	if err != nil {
		errChan <- fmt.Errorf("error reading file %s: %w", fp, err)
		return
	}

	err = json.Unmarshal(content, &res)
	if err != nil {
		errChan <- fmt.Errorf("error decoding JSON from file %s: %w", fp, err)
		return
	}

	dataChan <- res
}
