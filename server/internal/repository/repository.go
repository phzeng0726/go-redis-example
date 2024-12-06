package repository

import (
	"context"

	"github.com/phzeng0726/go-server-template/internal/domain"
	"github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

type Users interface {
	CreateUser(ctx context.Context, user domain.User) error
	GetUserByEmail(ctx context.Context, email string) (domain.User, error)
}

type Repositories struct {
	Users Users
}

func NewRepositories(db *gorm.DB, rdb *redis.Client) *Repositories {
	return &Repositories{
		Users: NewUsersRepo(db, rdb),
	}
}
