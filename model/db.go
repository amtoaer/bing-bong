package model

import (
	"fmt"

	"github.com/amtoaer/bing-bong/utils"
	"github.com/spf13/viper"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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
		utils.Fatal("不支持的数据库类型")
	}
	db, err = gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		utils.Fatalf("打开数据库失败：%v", err)
	}
	db.AutoMigrate(&User{}, &Feed{}, &Summary{})
	if err != nil {
		utils.Fatalf("初始化数据库失败：%v", err)
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
	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&Summary{Hash: hashStr}).Error
}

// 查询某人订阅的feeds
func QueryFeed(account int64, isGroup bool) (result []*Feed, err error) {
	user := &User{
		Account: account,
		IsGroup: isGroup,
	}
	db.Where(user).FirstOrCreate(user)
	err = db.Model(&user).Association("Feeds").Find(&result)
	return
}

// 查询带有订阅关系的用户列表
func QueryUser() (users []User, err error) {
	err = db.Preload("Feeds").Find(&users).Error
	return
}

func Search() (feeds []Feed) {
	db.Table("feeds").Find(&feeds)
	return
}

// 插入订阅关系
func InsertSubscription(url, name string, account int64, isGroup bool) error {
	user := &User{
		Account: account,
		IsGroup: isGroup,
	}
	feed := &Feed{
		URL:  url,
		Name: name,
	}
	db.Where(user).FirstOrCreate(user)
	db.Clauses(clause.OnConflict{ // 插入订阅时有可能出现链接对应的网站名称变更，冲突时仅更新网站名称
		DoUpdates: clause.AssignmentColumns([]string{"name"}),
	}).Where(feed).FirstOrCreate(feed)
	return db.Model(user).Association("Feeds").Append(feed)
}

// 删除订阅关系
func DeleteSubscription(url string, account int64, isGroup bool) error {
	user := &User{
		Account: account,
		IsGroup: isGroup,
	}
	feed := &Feed{
		URL: url,
	}
	db.Where(user).FirstOrCreate(user)
	db.Where(feed).FirstOrCreate(feed)
	return db.Model(user).Association("Feeds").Delete(feed)
}
