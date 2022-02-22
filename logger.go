package main

import (
	"go.uber.org/zap"
)

func getLogger() *zap.Logger {
	var config zap.Config
	if flagDev {
		config = zap.NewDevelopmentConfig()
	} else {
		config = zap.NewProductionConfig()
	}
	config.DisableStacktrace = true
	config.DisableCaller = true
	config.Level.SetLevel(flagLogLevel)
	logger, err := config.Build()
	if err != nil {
		panic("no logger: " + err.Error())
	}
	return logger
}
