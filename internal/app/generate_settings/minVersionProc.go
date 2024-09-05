package generateSettings

import (
	"cloud-terraform-mirror/internal/models"
	"strings"
)

func minVersionProc(versions []models.Version, minVersion string) ([]models.Version, error) {
	res := make([]models.Version, 0)

	for _, version := range versions {
		if isVersionGreaterOrEqual(version.Version, minVersion) {
			res = append(res, version)
		}
	}

	return res, nil
}

// compare versions strings
func isVersionGreaterOrEqual(v1, v2 string) bool {
	v1Parts := strings.Split(v1, ".")
	v2Parts := strings.Split(v2, ".")

	// Определяем максимальную длину, чтобы сравнить все части версии
	maxParts := len(v1Parts)
	if len(v2Parts) > maxParts {
		maxParts = len(v2Parts)
	}

	for i := 0; i < maxParts; i++ {
		var v1Part, v2Part string

		if i < len(v1Parts) {
			v1Part = v1Parts[i]
		}
		if i < len(v2Parts) {
			v2Part = v2Parts[i]
		}

		if v1Part > v2Part {
			return true
		} else if v1Part < v2Part {
			return false
		}
	}

	// Если все части версий равны
	return true
}
