package common

import (
	"flag"
	"moduleab_agent/conf"
)

var (
	Server   string
	LoginKey string
)

func init() {
	flag.StringVar(
		&Server, "server", "",
		"Specify ModuleAB server",
	)
	flag.StringVar(
		&LoginKey, "key", "",
		"Specify ModuleAB login key",
	)
	flag.Parse()
	if Server == "" {
		Server = conf.AppConfig.GetString("server")
	}
	if LoginKey == "" {
		LoginKey = conf.AppConfig.GetString("loginkey")
	}
}
