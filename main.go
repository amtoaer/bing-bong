package main

import (
	"github.com/amtoaer/bing-bong/client"
	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/model"
	"github.com/spf13/viper"
)

type robot interface {
	Login(string, string)
	SendMessage(int64, string, bool)
	HandleEvent(*message.MessageQueue)
}

func main() {
	var (
		bot robot               //机器人实例
		mq  = message.Default() //消息队列
	)
	// 根据机器人类型决定实例化内容
	switch viper.GetString("botType") {
	case "qq":
		bot = &client.QQ{}
	}
	// 登录机器人
	bot.Login(viper.GetString("botUser"), viper.GetString("botPass"))
	// 从数据库读取订阅信息初始化消息队列
	model.InitMessageQueue(bot, mq)
	// 启动机器人外部事件监听
	bot.HandleEvent(mq)
}
