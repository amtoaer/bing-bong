package model

import (
	"fmt"

	"github.com/amtoaer/bing-bong/utils"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var db *gorm.DB

// 初始化数据库链接
func InitDB() {
	var (
		dialector gorm.Dialector
		err       error
	)
	switch viper.GetString("dbType") {
	case "mysql":
		{
			conf := viper.GetStringMapString("mysql")
			address := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4",
				conf["dbUser"], conf["dbPass"],
				conf["dbAddress"], conf["dbName"])
			dialector = mysql.Open(address)
		}
	case "sqlite":
		{
			conf := viper.GetStringMapString("sqlite")
			dialector = sqlite.Open(conf["path"])
		}
	default:
		utils.Fatal("unsupported db type")
	}
	db, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		utils.Fatal("failed to open database")
	}
	db.AutoMigrate(&User{}, &Feed{}, &Summary{})
	if err != nil {
		utils.Fatal("failed to migrate database")
	}
}

// 查询文章是否已经存在
func IsFeedExist(hashStr string) bool {
	var count int64
	db.Table("summaries").Where("hash = ?", hashStr).Count(&count)
	return count > 0
}

// 插入已经推送的hash
func InsertHash(hashStr string) error {
	return db.Create(&Summary{Hash: hashStr}).Error
}

// 查询某人订阅的feeds
func QueryFeed(account int64) (result []*Feed) {
	db.Table("users").Where("account = ?", account).Association("Feeds").Find(&result)
	return result
}

// 添加订阅
func InsertSubscription(url, name string, account int64, isGroup bool) error {
	return db.Where(&User{
		Account: account,
		IsGroup: isGroup,
	}).Association("Feeds").Append(&Feed{
		URL:  url,
		Name: name,
	})
}

// 删除订阅
func DeleteSubscription(url string, account int64, isGroup bool) {
	db.Where(&User{
		Account: account,
		IsGroup: isGroup,
	}).Association("Feeds").Delete(&Feed{URL: url})
}
