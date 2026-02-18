package _repository

import (
	"Practice3/internal/_repository/_postgres"
	"Practice3/internal/_repository/_postgres/users"
	"Practice3/pkg/modules"
)

type UserRepository interface {
	GetUsers() ([]modules.User, error)
}

type Repositories struct {
	UserRepository
}

func NewRepositories(db *_postgres.Dialect) *Repositories {
	return &Repositories{
		UserRepository: users.NewUserRepository(db),
	}
}
