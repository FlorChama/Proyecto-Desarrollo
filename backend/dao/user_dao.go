package dao

import (
	"ticketek-backend/domain"

	"gorm.io/gorm"
)

type UserDAO struct {
	db *gorm.DB
}

func NewUserDAO(db *gorm.DB) *UserDAO {
	return &UserDAO{db: db}
}

func (d *UserDAO) Create(user *domain.User) error {
	return d.db.Create(user).Error
}

func (d *UserDAO) FindByEmail(email string) (*domain.User, error) {
	var user domain.User
	err := d.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) FindByID(id uint) (*domain.User, error) {
	var user domain.User
	err := d.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (d *UserDAO) UpdatePassword(id uint, hashedPassword string) error {
	return d.db.Model(&domain.User{}).Where("id = ?", id).Update("password", hashedPassword).Error
}
