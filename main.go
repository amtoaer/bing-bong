package main

import (
	"github.com/amtoaer/bing-bong/client"
	"github.com/spf13/viper"
)

type robot interface {
	Login(string, string)
	HandleEvent()
}

func main() {
	var bot robot
	switch viper.GetString("botType") {
	case "qq":
		bot = &client.QQ{}
	}
	bot.Login(viper.GetString("account"), viper.GetString("password"))
	bot.HandleEvent()
}
