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

	DatabaseConfig
}

type DatabaseConfig struct {
	// db
	DBType      string
	DBHost      string
	DBPort      string
	DBUser      string
	DBPasswd    string
	DBName      string
	DBHeartbeat int

	// Options
	DBConnectTimeout int
	DBCharset        string
	DBTimezone       string

	// MySQL/MariaDB
	ParseTime bool

	// Oracle
	ServiceName string
	TNS         string

	// ODBC
	Driver string
	DSN    string
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

	config.DBType = cfgFile.Section("DB").Key("TYPE").MustString("")
	config.DBHost = cfgFile.Section("DB").Key("HOST").MustString("")
	config.DBPort = cfgFile.Section("DB").Key("PORT").MustString("")
	config.DBName = cfgFile.Section("DB").Key("NAME").MustString("")
	// USER에 ## 주석이 들어가는 경우
	// decodedUser, _ := base64.StdEncoding.DecodeString(cfgFile.Section("DB").Key("USER").MustString(""))
	// config.DBUser = string(decodedUser)
	config.DBUser = cfgFile.Section("DB").Key("USER").MustString("")
	config.DBPasswd = cfgFile.Section("DB").Key("PASSWD").MustString("")
	config.DBHeartbeat = cfgFile.Section("DB").Key("DBHEARTBEAT").MustInt(30)

	config.DBConnectTimeout = cfgFile.Section("DB").Key("TIMEOUT").MustInt(10)
	config.DBCharset = cfgFile.Section("DB").Key("CHARSET").MustString("")
	config.DBTimezone = cfgFile.Section("DB").Key("TIMEZONE").MustString("")

	config.ParseTime = cfgFile.Section("DB").Key("PARSETIME").MustBool(true)

	config.ServiceName = cfgFile.Section("DB").Key("SERVICENAME").MustString("")
	config.TNS = cfgFile.Section("DB").Key("TNS").MustString("")

	config.Driver = cfgFile.Section("DB").Key("DRIVER").MustString("")
	config.DSN = cfgFile.Section("DB").Key("DSN").MustString("")

	return config
}
