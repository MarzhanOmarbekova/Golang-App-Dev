package usecase

import (
	"context"
	"fmt"
	"practice-7/internal/entity"
	"practice-7/internal/usecase/repo"
	"practice-7/utils"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

type UserUseCase struct {
	repo *repo.UserRepo
	rdb  *redis.Client
}

func NewUserUseCase(r *repo.UserRepo, rdb *redis.Client) *UserUseCase {
	return &UserUseCase{repo: r, rdb: rdb}
}

func (u *UserUseCase) RegisterUser(user *entity.User) (*entity.User, string, error) {

	code := utils.GenerateVerifyCode()
	user.VerifyCode = code
	user.Verified = false

	createdUser, err := u.repo.RegisterUser(user)
	if err != nil {
		return nil, "", fmt.Errorf("register user: %w", err)
	}

	go func() {
		if sendErr := utils.SendVerificationEmail(createdUser.Email, code); sendErr != nil {
			fmt.Printf("[EMAIL ERROR] %v\n", sendErr)
		}
	}()

	sessionID := uuid.New().String()
	return createdUser, sessionID, nil
}

func (u *UserUseCase) LoginUser(user *entity.LoginUserDTO) (map[string]string, error) {
	userFromRepo, err := u.repo.LoginUser(user)
	if err != nil {
		return nil, fmt.Errorf("User From Repo: %w", err)
	}

	if !utils.CheckPassword(userFromRepo.Password, user.Password) {
		return nil, fmt.Errorf("invalid credentials")
	}

	if !userFromRepo.Verified {
		return nil, fmt.Errorf("email not verified. Please verify your email first")
	}

	tokenPair, err := utils.GenerateJWT(userFromRepo.ID, userFromRepo.Role)
	if err != nil {
		return nil, fmt.Errorf("Generate JWT: %w", err)
	}

	if u.rdb != nil {
		ctx := context.Background()
		u.rdb.Set(ctx, "auth:"+userFromRepo.ID.String(), tokenPair.AccessToken, time.Minute*15)
	}

	return map[string]string{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	}, nil
}

func (u *UserUseCase) GetMe(userID uuid.UUID) (*entity.User, error) {
	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("GetMe: %w", err)
	}
	return user, nil
}

func (u *UserUseCase) PromoteUser(targetUserID uuid.UUID) (*entity.User, error) {
	user, err := u.repo.GetUserByID(targetUserID)
	if err != nil {
		return nil, fmt.Errorf("PromoteUser - get: %w", err)
	}
	user.Role = "admin"
	updated, err := u.repo.UpdateUser(user)
	if err != nil {
		return nil, fmt.Errorf("PromoteUser - update: %w", err)
	}
	return updated, nil
}

func (u *UserUseCase) VerifyEmail(username string, code string) error {
	loginDTO := &entity.LoginUserDTO{Username: username, Password: ""}
	user, err := u.repo.LoginUser(loginDTO)
	if err != nil {
		return fmt.Errorf("VerifyEmail - user not found: %w", err)
	}
	if user.VerifyCode != code {
		return fmt.Errorf("invalid verification code")
	}
	user.Verified = true
	user.VerifyCode = ""
	_, err = u.repo.UpdateUser(user)
	return err
}

func (u *UserUseCase) RefreshToken(refreshToken string) (map[string]string, error) {
	claims, err := utils.ParseJWT(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	tokenType, _ := claims["type"].(string)
	if tokenType != "refresh" {
		return nil, fmt.Errorf("not a refresh token")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid claims")
	}
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return nil, fmt.Errorf("invalid user_id")
	}

	user, err := u.repo.GetUserByID(userID)
	if err != nil {
		return nil, fmt.Errorf("user not found")
	}

	tokenPair, err := utils.GenerateJWT(userID, user.Role)
	if err != nil {
		return nil, err
	}

	if u.rdb != nil {
		ctx := context.Background()
		u.rdb.Set(ctx, "auth:"+userID.String(), tokenPair.AccessToken, time.Minute*15)
	}

	return map[string]string{
		"access_token":  tokenPair.AccessToken,
		"refresh_token": tokenPair.RefreshToken,
	}, nil
}

func (u *UserUseCase) LogoutUser(userID uuid.UUID) error {
	if u.rdb != nil {
		ctx := context.Background()
		return u.rdb.Del(ctx, "auth:"+userID.String()).Err()
	}
	return nil
}
