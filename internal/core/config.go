package core

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

type Web struct {
	Port string
}
type Control struct {
	Port       string
	BaudRate   int   `mapstructure:"baud_rate"`
	Resolution []int `mapstructure:"resolution"`
}
type Config struct {
	WebServer            Web              `mapstructure:"web"`
	InitUrl              string           `mapstructure:"init_url"`
	BaseUrl              string           `mapstructure:"base_url"`
	AssistPartyMemberMap map[string][]int `mapstructure:"assist_party_member_map"`
	Control
}

func InitConfig() (*Config, error) {
	cnfFile := flag.String("config", "configs/main.yaml", "config name")
	flag.Parse()
	fmt.Println(*cnfFile)
	viper.SetConfigFile(*cnfFile)
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
