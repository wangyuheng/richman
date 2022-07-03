package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfg = &Config{}

const (
	LogLevel              = "LOG_LEVEL"
	LarkAppId             = "LARK_APP_ID"
	LarkAppSecret         = "LARK_APP_SECRET"
	LarkAppToken          = "LARK_APP_TOKEN"
	LarkVerificationToken = "LARK_APP_VERIFICATION_TOKEN"
)

type Config struct {
	LogLevel logrus.Level
	LarkConfig
}

type LarkConfig struct {
	AppId             string
	AppSecret         string
	AppToken          string
	VerificationToken string
}

func Load() *Config {
	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault(LogLevel, logrus.InfoLevel.String())

	_ = v.BindEnv(LarkAppId)
	_ = v.BindEnv(LarkAppSecret)
	_ = v.BindEnv(LarkAppToken)
	_ = v.BindEnv(LarkVerificationToken)

	cfg.LarkConfig.AppId = v.GetString(LarkAppId)
	cfg.LarkConfig.AppSecret = v.GetString(LarkAppSecret)
	cfg.LarkConfig.AppToken = v.GetString(LarkAppToken)
	cfg.LarkConfig.VerificationToken = v.GetString(LarkVerificationToken)
	if l, err := logrus.ParseLevel(v.GetString(LogLevel)); err == nil {
		cfg.LogLevel = l
	}

	return cfg
}

func GetConfig() *Config {
	return cfg
}
