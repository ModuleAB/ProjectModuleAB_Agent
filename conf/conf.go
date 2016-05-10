package conf

import (
	"fmt"
	"io/ioutil"
	"moduleab_agent/consts"
	"os"
	"strings"
)

type Config map[string]interface{}

func ReadConfig(filename string) (Config, error) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	c := make(Config)
	c.parse(b)
	return c, nil
}

func (c Config) parse(b []byte) {
	lines := strings.Split(
		strings.Replace(string(b), "\r", "", -1),
		"\n",
	)
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		cfgs := strings.SplitN(line, "=", 2)
		c[cfgs[0]] = cfgs[1]
	}
}

func (c Config) Get(key string) (interface{}, error) {
	v, ok := c[key]
	if !ok {
		return nil, fmt.Errorf("Key", key, "not found.")
	}
	return v, nil
}

func (c Config) GetInt(key string) (int, error) {
	v, err := c.Get(key)
	if err != nil {
		return -1, err
	}
	i, ok := v.(int)
	if !ok {
		return -1, fmt.Errorf("Cannot convert to int.")
	}
	return i, nil
}

func (c Config) GetString(key string) string {
	v, _ := c.Get(key)
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return s
}

var AppConfig Config

func init() {
	var err error
	AppConfig, err = ReadConfig(consts.DefaultConfigFile)
	if err != nil {
		fmt.Fprintln(
			os.Stderr, consts.ErrorFormat,
			"Cannot read config file.", err,
		)
		os.Exit(1)
	}
}
