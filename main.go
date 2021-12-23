package main

import (
	"github.com/amtoaer/bing-bong/client"
	"github.com/amtoaer/bing-bong/internal"
	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/model"
	"github.com/amtoaer/bing-bong/utils"
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
	switch viper.GetString("botType") { // 根据机器人类型决定实例化内容
	case "qq":
		botConf = viper.GetStringMap("qq")
		bot = &client.QQ{Conf: botConf}
	}
	utils.InitLog()              // 重定向日志到文件
	bot.Init()                   // 登录机器人
	model.InitDB()               // 初始化数据库链接
	mm.Init(bot)                 // 从数据库读取订阅信息初始化消息队列，同时启动消息监听
	go internal.CheckMessage(mm) // 开始检测rss更新
	bot.HandleEvent(mm)          // 启动机器人外部事件监听
}
