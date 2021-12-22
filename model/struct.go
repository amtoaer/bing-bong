package model

type Summary struct {
	ID   uint
	Hash string `gorm:"index;unique"`
}

type User struct {
	ID      uint
	Account int64   `gorm:"uniqueIndex:idx_user;not null"`
	IsGroup bool    `gorm:"uniqueIndex:idx_user;not null"`
	Feeds   []*Feed `gorm:"many2many:subscriptions"`
}

type Feed struct {
	ID   uint
	URL  string `gorm:"index:idx_feed;unique;not null"`
	Name string `gorm:"index:idx_feed;not null"`
}
