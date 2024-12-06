package repository

import (
	"context"

	"github.com/phzeng0726/go-server-template/internal/domain"

	"gorm.io/gorm"
)

type Users interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type Repositories struct {
	Users Users
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		Users: NewUsersRepo(db),
	}
}
