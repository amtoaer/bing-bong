package model

import (
	"fmt"

	"github.com/spf13/viper"
)

func init() {
	// 读取配置文件
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	if err := viper.ReadInConfig(); err != nil {
		panic(fmt.Errorf("error reading config file: %v", err))
	}
	// 初始化数据库连接
	initDB()
}
