package model

import (
	"github.com/amtoaer/bing-bong/utils"
	"github.com/spf13/viper"
)

// 使用viper读取配置文件
func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		utils.Fatalf("读取配置文件失败：%v", err)
	}
}
