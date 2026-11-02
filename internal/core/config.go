package core

import (
	"flag"
	"fmt"

	"github.com/spf13/viper"
)

const (
	ServerTypeScryde string = "scryde"
)

type Web struct {
	Port string
}
type Control struct {
	Port       string
	BaudRate   int   `mapstructure:"baud_rate"`
	Resolution []int `mapstructure:"resolution"`
}
type Telegram struct {
	ApiKey string `mapstructure:"api_key"`
	ChatID int64  `mapstructure:"chat_id"`
}
type Config struct {
	ServerType           string           `mapstructure:"server_type"`
	WebServer            Web              `mapstructure:"web"`
	InitUrl              string           `mapstructure:"init_url"`
	BaseUrl              string           `mapstructure:"base_url"`
	AssistPartyMemberMap map[string][]int `mapstructure:"assist_party_member_map"`
	Control
	Telegram
}

func InitConfig() (*Config, error) {
	cName := flag.String("config", "configs/main.yaml", "config file")
	flag.Parse()
	fmt.Println(*cName)
	viper.SetConfigFile(*cName)
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
