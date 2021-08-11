package model

import (
	"database/sql"
	"fmt"

	"github.com/spf13/viper"
)

const (
	queryHashSQL  = "select count(*) from feeds where hash = ?"
	insertFeedSQL = "insert into feeds (url,hash) values (?,?)"
)

var (
	innerDB               *sql.DB
	queryHash, insertFeed *sql.Stmt
)

func initDB() {
	var err error
	address := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4",
		viper.GetString("user"), viper.GetString("password"),
		viper.GetString("address"), viper.GetString("database"))
	innerDB, err = sql.Open("mysql", address)
	if err != nil {
		panic(fmt.Errorf("error opening db connection:%v", err))
	}
	if err = innerDB.Ping(); err != nil {
		panic(fmt.Errorf("error ping db connection:%v", err))
	}
	queryHash, err = innerDB.Prepare(queryHashSQL)
	if err != nil {
		panic(fmt.Errorf("error constructing prepared statement:%v", err))
	}
	insertFeed, err = innerDB.Prepare(insertFeedSQL)
	if err != nil {
		panic(fmt.Errorf("error constructing prepared statement:%v", err))
	}
}

func IsFeedExist(hash string) bool {
	var count int
	queryHash.QueryRow(hash).Scan(&count)
	return count > 0
}

func InsertFeed(url, hash string) error {
	_, err := insertFeed.Exec(url, hash)
	return err
}
