package model

import (
	"fmt"

	"github.com/spf13/viper"
)

// 使用viper读取配置文件
func init() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config file: %v", err))
	}
	InitDB()
}
