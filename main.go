package main

import (
	"github.com/amtoaer/bing-bong/client"
	"github.com/amtoaer/bing-bong/internal"
	"github.com/amtoaer/bing-bong/message"
	"github.com/spf13/viper"
)

type robot interface {
	Init()
	SendMessage(int64, string, bool)
	HandleEvent(*message.Manager)
}

func main() {
	var (
		bot     robot                      //机器人实例
		mm      = message.DefaultManager() //消息队列
		botConf map[string]interface{}     //机器人配置
	)
	// 根据机器人类型决定实例化内容
	switch viper.GetString("botType") {
	case "qq":
		botConf = viper.GetStringMap("qq")
		bot = &client.QQ{Conf: botConf}
	}
	// 登录机器人
	bot.Init()
	// 从数据库读取订阅信息初始化消息队列
	mm.Init(bot)
	// 开始检测rss更新
	go internal.CheckMessage(mm)
	// 启动机器人外部事件监听
	bot.HandleEvent(mm)
}
