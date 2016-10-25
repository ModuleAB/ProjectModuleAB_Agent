package main

import (
	"fmt"
	"io/ioutil"
	"moduleab_agent/client"
	"moduleab_agent/common"
	"moduleab_agent/conf"
	"moduleab_agent/logger"
	"moduleab_agent/process"
	"moduleab_server/models"
	"moduleab_server/version"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"time"
)

func daemonStop() {
	bpid, err := ioutil.ReadFile(
		conf.AppConfig.GetString("pidfile"),
	)
	if err != nil {
		fmt.Println("Cannot find pid file, will not run.")
		os.Exit(1)
	}
	pid, err := strconv.Atoi(string(bpid))
	if err != nil {
		fmt.Println("Invalid pid, will not run.")
		os.Exit(1)
	}
	proc, err := os.FindProcess(pid)
	if err != nil {
		fmt.Println("Cannot fild proc=", pid)
		os.Exit(1)
	}
	err = proc.Kill()
	if err != nil {
		fmt.Println("Cannot stop daemon.")
		os.Exit(1)
	}
}

func showHelp() {
	fmt.Println("ModuleAB agent help")
	fmt.Println("\tUsage:", os.Args[0], "[stop|restart|help]")
	fmt.Println("\tDefault will start ModuleAB agent as daemon.")
	fmt.Println("\t\tstop: Stop the daemon.")
	fmt.Println("\t\trestart: Restart the daemon.")
	fmt.Println("\t\thelp: Show this help.")
}

func main() {
	timeout, err := conf.AppConfig.GetInt("timeout")
	if err == nil {
		client.StdHttp.Timeout = time.Duration(timeout) * time.Second
	}

	if len(os.Args[1:]) != 0 {
		switch os.Args[1] {
		case "stop":
			daemonStop()
			os.Exit(0)
		case "restart":
			fallthrough
		case "reload":
			daemonStop()
		default:
			showHelp()
			os.Exit(1)
		}
	}

	if os.Getppid() != 1 {
		exePath, _ := filepath.Abs(os.Args[0])
		cmd := exec.Command(exePath, os.Args[1:]...)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Start()
		fmt.Println("ModuleAB agent will run as daemon.")
		os.Exit(0)
	}

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

	var (
		d   *models.Hosts
		err error
	)
	for {
		d, err = client.RegisterHost()
		if err != nil {
			logger.AppLog.Debug("Got Error:", err)
			os.Exit(1)
		}
		if d == nil {
			logger.AppLog.Info("Register host succeed. waiting complete info.")
			fmt.Println("Register host succeed. waiting complete info.")
			time.Sleep(5 * time.Second)
			continue
		}
		if d.AppSet == nil {
			logger.AppLog.Info("App set not found. wait until ok.")
			fmt.Println("App set not found. wait until ok.")
			time.Sleep(5 * time.Second)
			continue
		}
		if len(d.Paths) == 0 {
			logger.AppLog.Info("No valid Path found. wait until ok.")
			fmt.Println("No valid Path found. wait until ok.")
			time.Sleep(5 * time.Second)
			continue
		}
		break
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
			logger.AppLog.Error("Connection to server is closed, reconnect.")
			time.Sleep(5 * time.Second)
		}
	}()
	b, err := process.NewBackupManager(*c, conf.AppConfig.GetBool("lowmemorymode"))
	if err != nil {
		logger.AppLog.Warn("Got error while making backup manager:", err)
		fmt.Println("Got error while making backup manager:", err)
		os.Exit(1)
	}
	logger.AppLog.Info("Starting backup manager...")
	b.Update(d.Paths)
	b.Run(d)
}
