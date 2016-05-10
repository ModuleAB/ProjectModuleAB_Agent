package main

import (
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/conf"
	"moduleab_agent/logger"
	"os"
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
		logger.AppLog.Debug("Got Error", err)
		os.Exit(1)
	}
	logger.AppLog.Debug("Got config", c.ApiKey, c.ApiSecret)
}
