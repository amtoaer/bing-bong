package model

type Summary struct {
	ID   uint
	Hash string `gorm:"index"`
}

type User struct {
	ID      uint
	Account int64
	IsGroup bool
	Feeds   []Feed `gorm:"many2many:subscriptions"`
}

type Feed struct {
	ID   uint
	URL  string
	Name string
}
