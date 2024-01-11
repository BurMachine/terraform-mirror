package main

import (
	createMirror "cloud-terraform-mirror/internal/app/create_mirror"
	generateSettings "cloud-terraform-mirror/internal/app/generate_settings"
	"cloud-terraform-mirror/internal/config"
	loggerLogrus "cloud-terraform-mirror/pkg/logger"
	"fmt"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	signalCh := make(chan os.Signal)
	signal.Notify(signalCh, os.Interrupt, os.Kill, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)

	logger := loggerLogrus.Init()

	conf := config.New()
	err := conf.LoadConfig()
	if err != nil {
		logger.Logger.Fatal(fmt.Sprintf("config loading error: %v", err.Error()))
	}

	exitChan := make(chan struct{})
	defer close(exitChan)

	go func() {
		err = generateSettings.Run(conf, logger)
		if err != nil {
			logger.Logger.Errorf("generate_settings error: %v", err.Error())
			signalCh <- syscall.SIGQUIT
		}
		createMirror.Run(conf, logger, exitChan)

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

	//db.Conn.Close(context.Background())
	logger.Logger.Info("service stop...")
}
