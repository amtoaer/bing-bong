package internal

import (
	"net/http"
	"net/url"
	"time"

	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/model"
	"github.com/amtoaer/bing-bong/utils"
	"github.com/mmcdole/gofeed"
	"github.com/spf13/viper"
)

var fp *gofeed.Parser = gofeed.NewParser()

func init() {
	proxyUrl := viper.GetString("proxy")
	if proxy, err := url.ParseRequestURI(proxyUrl); err == nil {
		fp.Client = &http.Client{Transport: &http.Transport{Proxy: http.ProxyURL(proxy)}}
	}
}

// 检测rss更新的定时任务
func CheckMessage(mq *message.MessageQueue) {
	checkTime := viper.GetInt64("checktime")
	for range time.Tick(time.Duration(checkTime) * time.Minute) {
		urls := mq.GetUrls()
		for _, url := range urls {
			go checkMessage(url, mq)
		}
	}
}

func checkMessage(url string, mq *message.MessageQueue) {
	feeds, err := fp.ParseURL(url)
	if utils.Errorf("error parsing urls: %v", err) {
		return
	}
	checkRange := min(feeds.Len(), viper.GetInt("checkrange")) //限制检测条数
	for i := 0; i < checkRange; i++ {
		feed := feeds.Items[i]
		hash := utils.Hash(feed)
		if !model.IsFeedExist(hash) {
			mq.Publish(url, utils.BuildMessage(feeds.Title, feed))
			model.InsertHash(url, hash)
		}
	}
}

func ParseTitle(url string) (string, error) {
	feeds, err := fp.ParseURL(url)
	return feeds.Title, err
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
