package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

// 取得feed条目的哈希值，防止重复推送
func Hash(item *gofeed.Item) string {
	var names []string
	for _, author := range item.Authors {
		names = append(names, author.Name)
	}
	return fmt.Sprintf("%X", sha256.Sum256(
		[]byte(item.Title+strings.Join(names, ","))))
}
