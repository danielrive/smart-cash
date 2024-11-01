package service

import (
	"log/slog"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
)

type UserService struct {
	userRepository *repositories.DynamoDBUsersRepository
	logger         *slog.Logger
}

var jwtKey = []byte("123456")

type claims = struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func NewUserService(userRepository *repositories.DynamoDBUsersRepository, logger *slog.Logger) *UserService {
	return &UserService{
		userRepository: userRepository,
		logger:         logger,
	}
}

func (us *UserService) GetUserById(userId string) (models.UserResponse, error) {

	user, err := us.userRepository.GetUserById(userId)

	if err != nil {
		return models.UserResponse{}, err
	}

	return user, nil
}

func (us *UserService) GetUserByEmailorUsername(key string, value string) (models.User, error) {
	// Find user
	user, err := us.userRepository.GetUserByEmailorUsername(key, value)

	if err != nil {
		return models.User{}, err
	}

	return user, err
}

func (us *UserService) CreateUser(u models.User) (models.UserResponse, error) {
	// generate UUID for the user

	user, err := us.userRepository.CreateUser(u)

	if err != nil {
		return models.UserResponse{}, err
	}

	return user, nil
}

// communicate with another service

func (us *UserService) Login(user string, password string) (string, error) {
	// validate password
	response, err := us.GetUserByEmailorUsername("username", user)

	if err != nil {
		return "", common.ErrWrongCredentials
	}
	if response.Password != password {
		us.logger.Error("authentication failed, wrong password",
			"username", user,
		)
		return "", common.ErrWrongCredentials

	}
	token, err := generateJWT(response.UserId)

	if err != nil {
		us.logger.Error("error generating token",
			"error", err.Error(),
			"username", user,
		)
		return "", common.ErrInternalError
	}

	return token, common.ErrWrongCredentials

}

func generateJWT(userID string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(jwtKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
