package usecase

import (
	"Practice3/internal/repository"
	"Practice3/pkg/modules"
)

type UserUsecase interface {
	GetUsers() ([]modules.User, error)
	GetUserByID(id int) (*modules.User, error)
	CreateUser(req modules.CreateUserRequest) (int, error)
	UpdateUser(id int, req modules.UpdateUserRequest) error
	DeleteUser(id int) (int64, error)
}

type Usecases struct {
	UserUsecase
}

func NewUsecases(repos *repository.Repositories) *Usecases {
	return &Usecases{
		UserUsecase: NewUserUsecase(repos.UserRepository),
	}
}
