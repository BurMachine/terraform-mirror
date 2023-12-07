package loggerLogrus

import "github.com/sirupsen/logrus"

type Logger struct {
	Logger *logrus.Logger
}

func Init() *Logger {
	logger := Logger{
		Logger: logrus.StandardLogger(),
	}

	logger.Logger.SetFormatter(&logrus.TextFormatter{
		DisableColors: true,
		FullTimestamp: true,
	})

	return &logger
}
