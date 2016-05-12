package main

import (
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/conf"
	"moduleab_agent/logger"
	"moduleab_agent/process"
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
		logger.AppLog.Debug("Got Error:", err)
		os.Exit(1)
	}
	logger.AppLog.Debug("Got config", c.ApiKey, c.ApiSecret)
	run(c)
}

func run(c *client.AliConfig) {
	d, err := client.RegisterHost()
	if err != nil {
		logger.AppLog.Debug("Got Error:", err)
		os.Exit(1)
	}
	if d == nil {
		logger.AppLog.Info("Register host succeed. waiting complete info.")
		os.Exit(0)
	}
	b, err := process.NewBackupManager(*c)
	if err != nil {
		logger.AppLog.Warn("Got error while making backup manager:", err)
		os.Exit(1)
	}
	if d.AppSet == nil {
		logger.AppLog.Info("App set not found. wait until ok.")
		os.Exit(1)
	}
	if len(d.Paths) == 0 {
		logger.AppLog.Info("No valid Path found. wait until ok.")
		os.Exit(1)
	}
	if len(d.ClientJobs) != 0 {
		r := process.NewRemoveManager()
		r.Update(d)
	}
	go func() {
		for {
			process.RunWebsocket(d, c.ApiKey, c.ApiSecret)
		}
	}()
	b.Update(d.Paths)
	b.Run(d)
}
