package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/phzeng0726/go-server-template/internal/domain"
	"github.com/redis/go-redis/v9"

	"gorm.io/gorm"
)

type UsersRepo struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewUsersRepo(db *gorm.DB, rdb *redis.Client) *UsersRepo {
	return &UsersRepo{
		db:  db,
		rdb: rdb,
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

	// Step 1: Try to get the user from Redis cache
	cacheKey := fmt.Sprintf("user:%s", email)
	cachedUser, err := r.rdb.Get(ctx, cacheKey).Result()
	if err == nil {
		// If the user exists in the cache, return the cached result
		// Assuming cachedUser is a JSON string, you might need to unmarshal it
		if err := json.Unmarshal([]byte(cachedUser), &user); err != nil {
			return user, fmt.Errorf("error unmarshalling cached data: %v", err)
		}
		return user, nil
	} else if err != redis.Nil {
		// Handle other Redis errors
		return user, fmt.Errorf("error checking cache: %v", err)
	}

	// Step 2: If not in cache, query the database
	if err := db.Where("email = ?", email).First(&user).Error; err != nil {
		return user, err
	}

	// Step 3: Cache the result in Redis (store in JSON format)
	go func() {
		userData, err := json.Marshal(user)
		if err != nil {
			// Use a channel or logging to report the error, instead of returning
			log.Printf("error marshalling user: %v", err)
			return
		}

		// Set the user data in Redis with an expiration time (e.g., 1 hour)
		if err := r.rdb.Set(ctx, cacheKey, userData, 1*time.Hour).Err(); err != nil {
			// Use a channel or logging to report the error, instead of returning
			log.Printf("error setting cache: %v", err)
			return
		}
	}()

	return user, nil
}
