package core

import (
	"github.com/spf13/viper"
)

type WebServer struct {
	Port string
}
type Config struct {
	WebServer WebServer `mapstructure:"web_server"`
}

func InitConfig() (*Config, error) {
	viper.SetConfigFile("configs/main.yaml")
	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}
	v := viper.New()
	v.SetConfigFile("configs/main.env.yaml")
	if v.ReadInConfig() == nil {
		err := viper.MergeConfigMap(v.AllSettings())
		if err != nil {
			return nil, err
		}
	}
	config := &Config{}
	if err := viper.Unmarshal(config); err != nil {
		return nil, err
	}

	return config, nil
}
