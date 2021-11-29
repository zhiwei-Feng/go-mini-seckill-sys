package domain

type Stock struct {
	ID      uint `gorm:"primary_key;"`
	Name    string
	Count   int
	Sale    int
	Version int
}

func (receiver Stock) TableName() string {
	return "stock"
}
