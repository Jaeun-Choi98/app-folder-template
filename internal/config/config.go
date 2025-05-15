package config

import (
	"log"

	"github.com/go-ini/ini"
)

type Configuration struct {
	RestIp   string
	RestPort string
}

func NewConfiguration() (*Configuration, error) {
	cfgFile, err := ini.Load("env.ini")
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return initConfig(cfgFile), nil
}

func initConfig(cfgFile *ini.File) *Configuration {
	config := &Configuration{}
	config.RestIp = cfgFile.Section("REST").Key("IP").MustString("")
	config.RestPort = cfgFile.Section("REST").Key("PORT").MustString("")
	return config
}
