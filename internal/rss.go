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
		fp.Client = &http.Client{
			Transport: &http.Transport{
				Proxy: http.ProxyURL(proxy),
			},
			Timeout: 15 * time.Second,
		}
	} else {
		fp.Client = &http.Client{
			Timeout: 15 * time.Second,
		}
	}
}

// 检测rss更新的定时任务
func CheckMessage(mm *message.Manager) {
	checkTime := viper.GetInt64("checktime")
	for range time.Tick(time.Duration(checkTime) * time.Minute) {
		urls := mm.GetTopics()
		for _, url := range urls {
			go checkMessage(url, mm)
		}
	}
}

func checkMessage(url string, mm *message.Manager) {
	feeds, err := fp.ParseURL(url)
	if err != nil {
		utils.Errorf("解析订阅链接失败：%v", err)
		return
	}
	checkRange := min(feeds.Len(), viper.GetInt("checkrange")) //限制检测条数
	for i := 0; i < checkRange; i++ {
		feed := feeds.Items[i]
		hash := utils.Hash(feed)
		if !model.IsFeedExist(hash) {
			mm.Publish(url, utils.BuildMessage(feeds.Title, feed))
			model.InsertHash(hash)
		}
	}
}

func ParseTitle(url string) (string, error) {
	feeds, err := fp.ParseURL(url)
	if err != nil {
		return "", err
	}
	checkRange := min(feeds.Len(), viper.GetInt("checkrange"))
	// 在订阅时自动读取已存在的条目，避免多次推送
	for i := 0; i < checkRange; i++ {
		model.InsertHash(utils.Hash(feeds.Items[i]))
	}
	return feeds.Title, err
}

func min(a, b int) int {
	if a > b {
		return b
	}
	return a
}
