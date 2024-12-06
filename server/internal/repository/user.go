package repository

import (
	"context"

	"github.com/phzeng0726/go-server-template/internal/domain"

	"gorm.io/gorm"
)

type UsersRepo struct {
	db *gorm.DB
}

func NewUsersRepo(db *gorm.DB) *UsersRepo {
	return &UsersRepo{
		db: db,
	}
}

func (r *UsersRepo) CreateUser(ctx context.Context, user domain.User) error {
	db := r.db.WithContext(ctx)

	if err := db.Create(&user).Error; err != nil {
		return err
	}

	return nil
}

func (r *UsersRepo) GetUserByEmail(ctx context.Context, email string) (domain.User, error) {
	var user domain.User
	db := r.db.WithContext(ctx)

	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return user, err
	}

	return user, nil
}
