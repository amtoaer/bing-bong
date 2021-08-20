package utils

import (
	"fmt"

	"github.com/mmcdole/gofeed"
)

// 消息模板
const tamplate = "您好，您关注的站点《%s》更新了新文章《%s》，快去看看吧！~\n链接：%s"

func BuildMessage(rssTitle string, item *gofeed.Item) string {
	return fmt.Sprintf(tamplate, rssTitle, item.Title, item.Link)
}
