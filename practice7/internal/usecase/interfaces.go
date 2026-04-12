package usecase

import (
	"practice-7/internal/entity"

	"github.com/google/uuid"
)

type UserInterface interface {
	RegisterUser(user *entity.User) (*entity.User, string, error)
	LoginUser(user *entity.LoginUserDTO) (map[string]string, error)
	GetMe(userID uuid.UUID) (*entity.User, error)
	PromoteUser(targetUserID uuid.UUID) (*entity.User, error)
	VerifyEmail(username string, code string) error
	RefreshToken(refreshToken string) (map[string]string, error)
	LogoutUser(userID uuid.UUID) error
}
