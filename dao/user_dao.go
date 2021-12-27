package dao

import (
	"gorm.io/gorm"
	"mini-seckill/domain"
)

func SelectUserByPk(db *gorm.DB, userId uint64) (domain.User, error) {
	user := domain.User{}
	result := db.First(&user, userId)
	if result.Error != nil {
		return user, result.Error
	}
	return user, nil
}
