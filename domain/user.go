package domain

type User struct {
	ID       uint64 `gorm:"primary_key;"`
	UserName string
}

func (u User) TableName() string {
	return "user"
}
