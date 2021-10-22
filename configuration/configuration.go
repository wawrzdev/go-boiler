package configuration

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	API_NAME string `json:"Name"`
	Server   *ServerConfiguration
	Database *DatabaseConfiguration
}

type ServerConfiguration struct {
	BIND_ADDRRESS string        `json:"BindAddress"`
	READ_TIMEOUT  time.Duration `json:"ReadTimeout"`
	WRITE_TIMEOUT time.Duration `json:"WriteTimeout"`
	IDLE_TIMEOUT  time.Duration `json:"IdleTimeout"`
}

type DatabaseConfiguration struct {
	DB_NAME     string `json:"DB_NAME"`
	DB_USER     string `json:"DB_USER"`
	DB_PASSWORD string `json:"DB_PASSWORD"`
}

func LoadConfiguration(name, fType string, filePaths *[]string) (config *Configuration, err error) {
	for _, v := range *filePaths {
		viper.AddConfigPath(v)
	}
	viper.SetConfigName(name)
	viper.SetConfigType(fType)
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		fmt.Println(err)
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			// config file was found but another error was produced
			return nil, err
		}
	}

	var c Configuration
	err = viper.Unmarshal(&c)
	if err != nil {
		return nil, err
	}

	return &c, nil
}

func SetDefaultConfiguration(defaultValues *map[string]interface{}) {
	for k, v := range *defaultValues {
		viper.SetDefault(k, v)
	}
}

func (sc *ServerConfiguration) GetServerConfiguration() (string, error) {
	scJSON, err := json.MarshalIndent(sc, "", "  ")
	if err != nil {
		return "", err
	}
	return string(scJSON), nil
}

func (dbc *DatabaseConfiguration) GetDatabaseConfiguration() (string, error) {
	dbcJSON, err := json.MarshalIndent(dbc, "", "  ")
	if err != nil {
		return "", nil
	}
	return string(dbcJSON), nil
}
