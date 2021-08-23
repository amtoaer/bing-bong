package model

import (
	"database/sql"
	"fmt"

	"github.com/amtoaer/bing-bong/message"
	"github.com/amtoaer/bing-bong/utils"
	"github.com/spf13/viper"

	_ "github.com/go-sql-driver/mysql"
)

const (
	queryHashSQL          = "select count(*) from hashes where hash = ?"
	insertHashSQL         = "insert ignore into hashes (url,hash) values (?,?)"
	querySubscriptionSQL  = "select url,subscriber,isGroup from subscriptions"
	insertSubscriptionSQL = "insert into subscriptions (url,subscriber,isGroup) values (?,?,?)"
	deleteSubscriptionSQL = "delete from subscriptions where url = ? and subscriber = ? and isGroup = ?"
	queryFeedSQL          = "select f.url,f.name from feeds f join subscriptions s on f.url=s.url and s.subscriber = ?"
	insertFeedSQL         = "replace into feeds (url,name) values (?,?)"
)

var (
	innerDB                                                   *sql.DB
	querySubscription, insertSubscription, deleteSubscription *sql.Stmt
	queryHash, insertHash, queryFeed, insertFeed              *sql.Stmt
)

type Feed struct {
	Url, Name string
}

// 初始化数据库链接
func InitDB() {
	var (
		err  error
		errs []error = make([]error, 7)
	)
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
	queryHash, errs[0] = innerDB.Prepare(queryHashSQL)
	insertHash, errs[1] = innerDB.Prepare(insertHashSQL)
	querySubscription, errs[2] = innerDB.Prepare(querySubscriptionSQL)
	insertSubscription, errs[3] = innerDB.Prepare(insertSubscriptionSQL)
	deleteSubscription, errs[4] = innerDB.Prepare(deleteSubscriptionSQL)
	queryFeed, errs[5] = innerDB.Prepare(queryFeedSQL)
	insertFeed, errs[6] = innerDB.Prepare(insertFeedSQL)
	utils.Fatalf("error constructing prepared statement:%v", errs...)
}

// 判断文章是否已经推送
func IsFeedExist(hash string) bool {
	var count int
	queryHash.QueryRow(hash).Scan(&count)
	return count > 0
}

// 插入已经推送的hash
func InsertHash(url, hash string) {
	_, err := insertHash.Exec(url, hash)
	utils.Errorf("error inserting feed:%v", err)
}

func QueryFeed(account int64) []*Feed {
	var (
		url, name string
		result    []*Feed
	)
	rows, err := queryFeed.Query(account)
	if utils.Errorf("error querying urls:%v", err) {
		return result
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&url, &name)
		result = append(result, &Feed{url, name})
	}
	return result
}

// 初始化消息队列（即查询订阅者并遍历订阅）
func InitMessageQueue(bot message.Messager, mq *message.MessageQueue) {
	var (
		url     string
		account int64
		isGroup bool
	)
	rows, err := querySubscription.Query()
	if utils.Errorf("error querying subscriber:%v", err) {
		return
	}
	defer rows.Close()
	for rows.Next() {
		rows.Scan(&url, &account, &isGroup)
		mq.Subscribe(bot, url, account, isGroup)
	}
}

// 插入订阅者
func InsertSubscription(url, name string, account int64, isGroup bool) {
	_, insertUrlErr := insertFeed.Exec(url, name)
	_, insertSubscriberErr := insertSubscription.Exec(url, account, isGroup)
	utils.Errorf("error inserting subscriber:%v", insertUrlErr, insertSubscriberErr)
}

func DeleteSubscription(url string, account int64, isGroup bool) {
	_, err := deleteSubscription.Exec(url, account, isGroup)
	utils.Errorf("error deleting subscription:%v", err)
}
