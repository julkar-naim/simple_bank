package util

import (
	"github.com/spf13/viper"
	"path/filepath"
	"runtime"
)

type Config struct {
	DBDriver string `mapstructure:"DB_DRIVER"`
	DBSource string `mapstructure:"DB_SOURCE"`
	Address  string `mapstructure:"ADDRESS"`
}

var AppConfig Config

func NewConfig(c Config) {
	AppConfig = c
}

func LoadConfig() (config Config, err error) {
	_, filename, _, _ := runtime.Caller(0)
	dir := filepath.Dir(filename)
	viper.AddConfigPath(dir + "/../")
	viper.SetConfigName("app")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err = viper.ReadInConfig()
	if err != nil {
		return
	}
	err = viper.Unmarshal(&config)
	return
}
