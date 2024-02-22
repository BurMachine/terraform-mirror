package main

import (
	createMirror "cloud-terraform-mirror/internal/app/create_mirror"
	generateSettings "cloud-terraform-mirror/internal/app/generate_settings"
	"cloud-terraform-mirror/internal/clean"
	"cloud-terraform-mirror/internal/config"
	"cloud-terraform-mirror/internal/obs_uploading"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"fmt"
	"github.com/sirupsen/logrus"
	"io"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	logger := loggerLogrus.Init()
	t := time.Now()
	logFileName := fmt.Sprintf("%s.log", t.Format(time.RFC3339))
	file, err := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND|os.O_RDWR, 0666)
	if err == nil {
		mw := io.MultiWriter(os.Stdout, file)
		logrus.SetOutput(mw)
	} else {
		logger.Logger.Fatal("logger init error")
	}

	mw := io.MultiWriter(os.Stdout, file)
	logrus.SetOutput(mw)

	err = clean.Clean(logFileName)
	if err != nil {
		logger.Logger.Fatal("cleaning error: ", err)
	}

	conf := config.New()
	err = conf.LoadConfig()
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("config loading error: %v", err.Error()))
	}

	exitChan := make(chan struct{})
	defer close(exitChan)

	errFlag := false
	go func() {
		err = generateSettings.Run(conf, logger)
		if err != nil {
			logger.Logger.Errorf("generate_settings error: %v", err.Error())
			signalCh <- syscall.SIGQUIT
			errFlag = true
			return
		}
		err = createMirror.Run(conf, logger, exitChan)
		if err != nil {
			logger.Logger.Errorf("create mirror error: %v", err.Error())
			signalCh <- syscall.SIGQUIT
			errFlag = true
			return
		}
		signalCh <- syscall.SIGQUIT
	}()

	sig, ok := <-signalCh
	if ok {
		if sig != syscall.SIGQUIT {
			logger.Logger.Info("gracefully stopping...")
			exitChan <- struct{}{}
			<-exitChan
		}
	}

	err = obs_uploading.ObsUploadLog(conf, logFileName, errFlag)
	if err != nil {
		logger.Logger.Error(err)
	}
	logger.Logger.Info("service stop...")
}
