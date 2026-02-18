package users

import (
	"Practice3/internal/_repository/_postgres"
	"Practice3/pkg/modules"
	"fmt"
	"time"
)

type Repository struct {
	db               *_postgres.Dialect
	executionTimeout time.Duration
}

func NewUserRepository(db *_postgres.Dialect) *Repository {
	return &Repository{
		db:               db,
		executionTimeout: time.Second * 5,
	}
}

func (r *Repository) GetUsers() ([]modules.User, error) {
	var users []modules.User
	err := r.db.DB.Select(&users, "SELECT id, name, email, age, created_at FROM users")
	if err != nil {
		return nil, err
	}
	fmt.Println(users)
	return users, nil
}
