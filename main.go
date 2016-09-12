package main

import (
	"fmt"
	"io/ioutil"
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/conf"
	"moduleab_agent/logger"
	"moduleab_agent/process"
	"moduleab_server/version"
	"os"
	"runtime"
	"time"
)

func main() {
	ioutil.WriteFile(
		conf.AppConfig.GetString("pidfile"),
		[]byte(fmt.Sprint(os.Getpid())),
		0600,
	)

	logger.Init()

	logger.AppLog.Info("ModuleAB agent", version.Version, "starting...")
	logger.AppLog.Level = logger.StringLevelToInt(
		conf.AppConfig.GetString("loglevel"),
	)
	logger.AppLog.Debug("Got server:", common.Server)
	logger.AppLog.Debug("Got login key:", common.LoginKey)
	c, err := client.GetAliConfig()
	if err != nil {
		logger.AppLog.Fatal("Got Error:", err)
		os.Exit(1)
	}
	logger.AppLog.Debug("Got config", c.ApiKey, c.ApiSecret)

	for {
		run(c)
		logger.AppLog.Error("Main thread crashed, restarting...")
	}
}

func run(c *client.AliConfig) {
	defer func() {
		x := recover()
		if x != nil {
			logger.AppLog.Error("Got fatal error:", x)
			var stack = make([]byte, 2<<10)
			runtime.Stack(stack, true)
			logger.AppLog.Error("Stack trace:\n", string(stack))
		}
	}()
	d, err := client.RegisterHost()
	if err != nil {
		logger.AppLog.Debug("Got Error:", err)
		os.Exit(1)
	}
	if d == nil {
		logger.AppLog.Info("Register host succeed. waiting complete info.")
		fmt.Println("Register host succeed. waiting complete info.")
		os.Exit(0)
	}
	b, err := process.NewBackupManager(*c)
	if err != nil {
		logger.AppLog.Warn("Got error while making backup manager:", err)
		fmt.Println("Got error while making backup manager:", err)
		os.Exit(1)
	}
	if d.AppSet == nil {
		logger.AppLog.Info("App set not found. wait until ok.")
		fmt.Println("App set not found. wait until ok.")
		os.Exit(1)
	}
	if len(d.Paths) == 0 {
		logger.AppLog.Info("No valid Path found. wait until ok.")
		fmt.Println("No valid Path found. wait until ok.")
		os.Exit(1)
	}
	logger.AppLog.Info("Starting remove manager...")
	if len(d.ClientJobs) != 0 {
		r := process.NewRemoveManager()
		r.Update(d)
	}
	logger.AppLog.Info("Starting recover manager...")
	go func() {
		for {
			process.RunWebsocket(d, c.ApiKey, c.ApiSecret)
			time.Sleep(5 * time.Second)
		}
	}()
	logger.AppLog.Info("Starting backup manager...")
	b.Update(d.Paths)
	b.Run(d)
}
