package generateSettings

import (
	"cloud-terraform-mirror/internal/models"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
)

// возвращать true/false есть тру то есть дельта, если в ффалсе, то файл просто скопирован в дельты чтобы create всегда читал оттуда
// , а потом удалял, ошибку возвращал только когда что-то сломалось
func checkExistVersions(bodyBytes []byte, outputDir, namespace, name string, minVersion string) (err error, settingsFilesEqual bool) {
	if _, err = os.Stat(outputDir); os.IsNotExist(err) {
		err = os.MkdirAll(outputDir, os.ModePerm)
		if err != nil {
			return
		}
	}

	// Check exist
	file, err := os.Open(fmt.Sprintf("%s/%s-%s.json", outputDir, namespace, name))
	if err != nil {
		if os.IsNotExist(err) {
			err = createDefaultDeltaFile(bodyBytes, outputDir, namespace, name)
			if err != nil {
				return
			}
			return
		}
		return
	}
	fi, err := file.Stat()
	if err != nil {
		return
	}
	if fi.Size() == 0 {
		err = createDefaultDeltaFile(bodyBytes, outputDir, namespace, name)
		if err != nil {
			return
		}
		return
	}

	// Check diff
	moduleNew := models.Module{}
	moduleOld := models.Module{}

	err = json.Unmarshal(bodyBytes, &moduleNew)
	if err != nil {
		return
	}

	decoder := json.NewDecoder(file)

	err = decoder.Decode(&moduleOld)
	if err != nil {
		return
	}

	versions := findNewVersions(moduleOld.Versions, moduleNew.Versions)

	resultModule := models.Module{
		ID:       moduleOld.ID,
		Versions: versions,
	}
	resBytes, err := json.Marshal(resultModule)
	if err != nil {
		return
	}
	if len(resultModule.Versions) == 0 {
		err = createDefaultDeltaFile(resBytes, outputDir, namespace, name)
		if err != nil {
			return
		}
	} else {
		err = createDefaultDeltaFile(resBytes, outputDir, namespace, name)
		if err != nil {
			return
		}
	}

	err = minVersionProcessing()

	return
}

func minVersionProcessing() error {

	return nil
}

func createDefaultDeltaFile(bodyBytes []byte, outDir, namespace, name string) error {
	if _, err := os.Stat(fmt.Sprintf("%s/deltas", outDir)); os.IsNotExist(err) {
		if err = os.MkdirAll(fmt.Sprintf("%s/deltas", outDir), os.ModePerm); err != nil {
			return err
		}
	}
	outputFilePath := filepath.Join(outDir, fmt.Sprintf("deltas/%s-%s-delta.json", namespace, name))
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

func findNewVersions(versions1, versions2 []models.Version) []models.Version {
	newVersions := make([]models.Version, 0)

	versionsMap := make(map[string]struct{})
	for _, v := range versions1 {
		versionsMap[v.Version] = struct{}{}
	}

	for _, v := range versions2 {
		if _, exists := versionsMap[v.Version]; !exists {
			newVersions = append(newVersions, v)
		}
	}

	sort.Slice(newVersions, func(i, j int) bool {
		return newVersions[i].Version < newVersions[j].Version
	})

	return newVersions
}
