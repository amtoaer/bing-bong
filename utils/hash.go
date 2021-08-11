package utils

import (
	"crypto/sha256"
	"fmt"
	"strings"

	"github.com/mmcdole/gofeed"
)

func Hash(item *gofeed.Item) string {
	var names []string
	for _, author := range item.Authors {
		names = append(names, author.Name)
	}
	return fmt.Sprintf("%X", sha256.Sum256(
		[]byte(item.Title+strings.Join(names, ","))))
}
