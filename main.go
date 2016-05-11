package main

import (
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/conf"
	"moduleab_agent/logger"
	"os"
	"time"
)

func main() {
	logger.AppLog.Info("ModuleAB agent", common.Version, "starting...")
	logger.AppLog.Level = logger.StringLevelToInt(
		conf.AppConfig.GetString("loglevel"),
	)
	logger.AppLog.Debug("Got server:", common.Server)
	logger.AppLog.Debug("Got login key:", common.LoginKey)
	c, err := client.GetAliConfig()
	if err != nil {
		logger.AppLog.Debug("Got Error:", err)
		os.Exit(1)
	}
	logger.AppLog.Debug("Got config", c.ApiKey, c.ApiSecret)
}

func run() {
	d, err := client.RegisterHost()
	if err != nil {
		logger.AppLog.Debug("Got Error:", err)
		os.Exit(1)
	}
	if d == nil {
		logger.AppLog.Info("Register host succeed. waiting complete info.")
	}
	for {
		select {
		case <-time.Tick(time.Minute):
			d, err := client.RegisterHost()
			if err != nil {
				logger.AppLog.Debug("Got Error:", err)
				continue
			}
			if d.AppSet == nil {
				logger.AppLog.Info("App set not found. wait until ok.")
				continue
			}
			if len(d.BackupSets) == 0 {
				logger.AppLog.Info("No valid backup set found. wait until ok.")
				continue
			}

		}
	}
}
