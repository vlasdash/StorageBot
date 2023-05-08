package config

import (
	"github.com/spf13/viper"
)

type Config struct {
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	WebhookURL              string `yaml:"webhook_url"`
	PasswordRetentionMinute int    `yaml:"password_retention_minute"`
	Port                    int    `yaml:"port"`
}

var C Config

func LoadConfig(path string) error {
	viper.AddConfigPath(path)
	viper.SetConfigName("config")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return err
	}
	C.App.PasswordRetentionMinute = viper.GetStringMap("app")["password_retention_minute"].(int)
	C.App.WebhookURL = viper.GetStringMap("app")["webhook_url"].(string)
	C.App.Port = viper.GetStringMap("app")["port"].(int)

	return nil
}
