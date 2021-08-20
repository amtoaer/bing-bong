package model

import (
	"database/sql"
	"fmt"

	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/utils"
	"github.com/spf13/viper"
)

const (
	queryHashSQL        = "select count(*) from feeds where hash = ?"
	insertFeedSQL       = "insert into feeds (url,hash) values (?,?)"
	querySubscriberSQL  = "select url,subscriber from maps"
	insertSubscriberSQL = "insert into maps (url,subscriber) values (?,?)"
)

var (
	innerDB                                                  *sql.DB
	queryHash, insertFeed, querySubscriber, insertSubscriber *sql.Stmt
)

// 初始化数据库链接
func InitDB() {
	var err, first, second, third, forth error
	address := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4",
		viper.GetString("dbUser"), viper.GetString("dbPass"),
		viper.GetString("dbAddress"), viper.GetString("dbName"))
	innerDB, err = sql.Open("mysql", address)
	if err != nil {
		panic(fmt.Errorf("error opening db connection:%v", err))
	}
	if err = innerDB.Ping(); err != nil {
		panic(fmt.Errorf("error ping db connection:%v", err))
	}
	queryHash, first = innerDB.Prepare(queryHashSQL)
	insertFeed, second = innerDB.Prepare(insertFeedSQL)
	querySubscriber, third = innerDB.Prepare(querySubscriberSQL)
	insertSubscriber, forth = innerDB.Prepare(insertSubscriberSQL)
	utils.CheckError("error constructing prepared statement:%v", first, second, third, forth)
}

// 判断文章是否已经推送
func IsFeedExist(hash string) bool {
	var count int
	queryHash.QueryRow(hash).Scan(&count)
	return count > 0
}

// 插入已经推送的hash
func InsertFeed(url, hash string) error {
	_, err := insertFeed.Exec(url, hash)
	return err
}

// 初始化消息队列（即查询订阅者并遍历订阅）
func InitMessageQueue(bot message.Messager, mq *message.MessageQueue) {
	var (
		url     string
		account int64
		isGroup bool
	)
	rows, err := querySubscriber.Query()
	utils.CheckError("error querying subscribers:%v", err)
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&url, &account, &isGroup)
		mq.Subscribe(bot, url, account, isGroup)
	}
}

// 插入订阅者
func InsertSubscriber(url string, account int64) error {
	_, err := insertSubscriber.Exec(url, account)
	return err
}
