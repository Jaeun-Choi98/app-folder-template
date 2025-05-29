package config

import (
	"pjt/internal/logger"
	"sync"

	"github.com/go-ini/ini"
)

type Configuration struct {
	// for saving to .ini file
	mu      sync.RWMutex
	cfgFile *ini.File

	// http-rest
	RestIp   string
	RestPort string

	// repository
	User   string
	Passwd string
	Conn   string
}

func NewConfiguration() (*Configuration, error) {
	cfgFile, err := ini.Load("env.ini")
	if err != nil {
		logger.Println(err)
		return nil, err
	}
	return initConfig(cfgFile), nil
}

func initConfig(cfgFile *ini.File) *Configuration {
	config := &Configuration{}

	config.cfgFile = cfgFile

	config.RestIp = cfgFile.Section("REST").Key("IP").MustString("")
	config.RestPort = cfgFile.Section("REST").Key("PORT").MustString("")
	config.User = cfgFile.Section("ORACLE").Key("USER").MustString("")
	config.Passwd = cfgFile.Section("ORACLE").Key("PASSWD").MustString("")
	config.Conn = cfgFile.Section("ORACLE").Key("CONN").MustString("")
	return config
}
