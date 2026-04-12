package repo

import (
	"fmt"
	"practice-7/internal/entity"
	"practice-7/pkg/postgres"

	"github.com/google/uuid"
)

type UserRepo struct {
	PG *postgres.Postgres
}

func NewUserRepo(pg *postgres.Postgres) *UserRepo {
	return &UserRepo{pg}
}

func (u *UserRepo) RegisterUser(user *entity.User) (*entity.User, error) {
	err := u.PG.Conn.Create(user).Error
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepo) LoginUser(user *entity.LoginUserDTO) (*entity.User, error) {
	var userFromDB entity.User
	if err := u.PG.Conn.Where("username = ?", user.Username).First(&userFromDB).Error; err != nil {
		return nil, fmt.Errorf("Username Not Found: %v", err)
	}
	return &userFromDB, nil
}

func (u *UserRepo) GetUserByID(userID uuid.UUID) (*entity.User, error) {
	var user entity.User
	if err := u.PG.Conn.Where("id = ?", userID).First(&user).Error; err != nil {
		return nil, fmt.Errorf("User Not Found: %v", err)
	}
	return &user, nil
}

func (u *UserRepo) UpdateUser(user *entity.User) (*entity.User, error) {
	if err := u.PG.Conn.Save(user).Error; err != nil {
		return nil, err
	}
	return user, nil
}

func (u *UserRepo) GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User
	if err := u.PG.Conn.Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("User Not Found: %v", err)
	}
	return &user, nil
}
