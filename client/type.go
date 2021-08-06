package client

var Robot Client

type Client interface {
	init()
	SendMessage(string, string)
}

func init() {
	// 初始化全局client
}
