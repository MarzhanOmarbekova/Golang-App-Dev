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
	RestoreUser(id int) error

	GetPaginatedUsers(page, pageSize int, filter modules.UserFilter, sort modules.UserSort) (modules.PaginatedResponse, error)
	GetCursorPaginatedUsers(cursor, limit int, filter modules.UserFilter) (modules.CursorPaginatedResponse, error)

	GetCommonFriends(user1ID, user2ID int) ([]modules.User, error)
	AddFriend(userID, friendID int) error
	GetFriendRecommendations(userID int) ([]modules.FriendRecommendation, error)
}

type Usecases struct {
	UserUsecase
}

func NewUsecases(repos *repository.Repositories) *Usecases {
	return &Usecases{
		UserUsecase: NewUserUsecase(repos.UserRepository),
	}
}
