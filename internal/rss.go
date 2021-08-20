package internal

import (
	"time"

	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/model"
	"github.com/amtoaer/bing-bong/utils"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

var fp gofeed.Parser

// 检测rss更新的定时任务
func CheckMessage() {
	checkTime := viper.GetInt64("checkTime")
	for range time.Tick(time.Duration(checkTime) * time.Minute) {
		urls := message.Default().GetUrls()
		mq := message.Default()
		for _, url := range urls {
			go checkMessage(url, mq)
		}
	}
}

func checkMessage(url string, mq *message.MessageQueue) {
	feeds, err := fp.ParseURL(url)
	utils.CheckError("error parsing urls: %v", err)
	checkRange := max(feeds.Len(), viper.GetInt("checkRange")) //限制检测条数
	for i := 0; i < checkRange; i++ {
		feed := feeds.Items[i]
		hash := utils.Hash(feed)
		if !model.IsFeedExist(hash) {
			mq.Publish(url, utils.BuildMessage(feed.Title, feed))
			model.InsertFeed(url, hash)
		}
	}
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
