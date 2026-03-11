package usecase

import (
	"Practice3/internal/repository"
	"Practice3/pkg/modules"
)

type userUsecase struct {
	repo repository.UserRepository
}

// NewUserUsecase returns a concrete UserUsecase backed by the given repository.
func NewUserUsecase(repo repository.UserRepository) UserUsecase {
	return &userUsecase{repo: repo}
}

func (u *userUsecase) GetUsers() ([]modules.User, error) {
	return u.repo.GetUsers()
}

func (u *userUsecase) GetUserByID(id int) (*modules.User, error) {
	return u.repo.GetUserByID(id)
}

func (u *userUsecase) CreateUser(req modules.CreateUserRequest) (int, error) {
	return u.repo.CreateUser(req)
}

func (u *userUsecase) UpdateUser(id int, req modules.UpdateUserRequest) error {
	return u.repo.UpdateUser(id, req)
}

func (u *userUsecase) DeleteUser(id int) (int64, error) {
	return u.repo.DeleteUser(id)
}

func (u *userUsecase) RestoreUser(id int) error {
	return u.repo.RestoreUser(id)
}

func (u *userUsecase) GetPaginatedUsers(
	page, pageSize int,
	filter modules.UserFilter,
	sort modules.UserSort,
) (modules.PaginatedResponse, error) {
	return u.repo.GetPaginatedUsers(page, pageSize, filter, sort)
}

func (u *userUsecase) GetCursorPaginatedUsers(
	cursor, limit int,
	filter modules.UserFilter,
) (modules.CursorPaginatedResponse, error) {
	return u.repo.GetCursorPaginatedUsers(cursor, limit, filter)
}

func (u *userUsecase) GetCommonFriends(user1ID, user2ID int) ([]modules.User, error) {
	return u.repo.GetCommonFriends(user1ID, user2ID)
}

func (u *userUsecase) AddFriend(userID, friendID int) error {
	return u.repo.AddFriend(userID, friendID)
}

func (u *userUsecase) GetFriendRecommendations(userID int) ([]modules.FriendRecommendation, error) {
	return u.repo.GetFriendRecommendations(userID)
}
