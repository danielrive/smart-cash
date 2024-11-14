package service

import (
	"context"
	"log/slog"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/models"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
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

func (us *UserService) GetUserById(ctx context.Context, userId string) (models.UserResponse, error) {
	tr := otel.Tracer("user-service")
	trContext, childSpan := tr.Start(ctx, "SVCGetUserById")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()
	user, err := us.userRepository.GetUserById(trContext, userId)

	if err != nil {
		return models.UserResponse{}, err
	}

	return user, nil
}

func (us *UserService) GetUserByEmailorUsername(ctx context.Context, key string, value string) (models.User, error) {
	tr := otel.Tracer("user-service")
	trContext, childSpan := tr.Start(ctx, "SVCGetUserByEmailorUsername")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	user, err := us.userRepository.GetUserByEmailorUsername(trContext, key, value)

	if err != nil {
		return models.User{}, err
	}

	return user, err
}

func (us *UserService) CreateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	tr := otel.Tracer("user-service")
	trContext, childSpan := tr.Start(ctx, "SVCCreateUser")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()

	user, err := us.userRepository.CreateUser(trContext, u)

	if err != nil {
		return models.UserResponse{}, err
	}

	return user, nil
}

// communicate with another service

func (us *UserService) Login(ctx context.Context, user string, password string) (string, error) {
	tr := otel.Tracer("user-service")
	trContext, childSpan := tr.Start(ctx, "SVCLogin")
	childSpan.SetAttributes(attribute.String("component", "service"))
	defer childSpan.End()
	response, err := us.GetUserByEmailorUsername(trContext, "username", user)

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
