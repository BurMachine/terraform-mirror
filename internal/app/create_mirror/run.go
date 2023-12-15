package createMirror

import (
	"cloud-terraform-mirror/internal/config"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
)

func Run(conf *config.Conf, logger *loggerLogrus.Logger) {
	logger.Logger.Info("starting creating local mirror")

}
