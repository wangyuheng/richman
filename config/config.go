package config

import (
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

var cfg = &Config{}

const (
	LogLevel             = "LOG_LEVEL"
	AiURL                = "AI_URL"
	AiKey                = "AI_KEY"
	LarkAppId            = "LARK_APP_ID"
	LarkAppSecret        = "LARK_APP_SECRET"
	WechatToken          = "WECHAT_TOKEN"
	TemplateAppToken     = "TEMPLATE_APP_TOKEN"
	TargetFolderAppToken = "TARGET_FOLDER_APP_TOKEN"
	DBAppToken           = "DB_APP_TOKEN"
	DBTableToken         = "DB_TABLE_TOKEN"
)

type Config struct {
	LogLevel logrus.Level
	LarkConfig
	AIConfig
	LarkDBConfig
}

type AIConfig struct {
	AiURL string
	AiKey string
}

type LarkConfig struct {
	DbAppId              string
	DbAppSecret          string
	WechatToken          string
	TemplateAppToken     string
	TargetFolderAppToken string
}

type LarkDBConfig struct {
	DBAppToken           string
	DBTableToken         string
	TemplateAppToken     string
	TargetFolderAppToken string
}

func Load() *Config {
	v := viper.New()
	v.AutomaticEnv()
	v.SetDefault(LogLevel, logrus.InfoLevel.String())

	_ = v.BindEnv(AiURL)
	_ = v.BindEnv(AiKey)
	_ = v.BindEnv(LarkAppId)
	_ = v.BindEnv(LarkAppSecret)
	_ = v.BindEnv(WechatToken)
	_ = v.BindEnv(TemplateAppToken)
	_ = v.BindEnv(TargetFolderAppToken)
	_ = v.BindEnv(DBAppToken)
	_ = v.BindEnv(DBTableToken)

	cfg.AIConfig.AiURL = v.GetString(AiURL)
	cfg.AIConfig.AiKey = v.GetString(AiKey)
	cfg.LarkConfig.DbAppId = v.GetString(LarkAppId)
	cfg.LarkConfig.DbAppSecret = v.GetString(LarkAppSecret)
	cfg.LarkConfig.WechatToken = v.GetString(WechatToken)
	cfg.LarkConfig.TemplateAppToken = v.GetString(TemplateAppToken)
	cfg.LarkConfig.TargetFolderAppToken = v.GetString(TargetFolderAppToken)
	cfg.LarkDBConfig.DBAppToken = v.GetString(DBAppToken)
	cfg.LarkDBConfig.DBTableToken = v.GetString(DBTableToken)
	cfg.LarkDBConfig.TemplateAppToken = v.GetString(TemplateAppToken)
	cfg.LarkDBConfig.TargetFolderAppToken = v.GetString(TargetFolderAppToken)
	if l, err := logrus.ParseLevel(v.GetString(LogLevel)); err == nil {
		cfg.LogLevel = l
	}

	return cfg
}

func GetConfig() *Config {
	return cfg
}

func GetLarkDBConfig() LarkDBConfig {
	return cfg.LarkDBConfig
}

func GetLarkConfig() LarkConfig {
	return cfg.LarkConfig
}
