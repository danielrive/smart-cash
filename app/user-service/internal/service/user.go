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
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository *repositories.DynamoDBUsersRepository
	logger         *slog.Logger
	jwtSecret      []byte
}

type claims = struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

func NewUserService(userRepository *repositories.DynamoDBUsersRepository, jwtSecret []byte, logger *slog.Logger) *UserService {
	return &UserService{
		userRepository: userRepository,
		logger:         logger,
		jwtSecret:      jwtSecret,
	}
}

func (us *UserService) GetUserById(ctx context.Context, userId string) (models.UserResponse, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetUserById")
	defer childSpan.End()
	user, err := us.userRepository.GetUserById(trContext, userId)

	if err != nil {
		return models.UserResponse{}, err
	}

	return user, nil
}

func (us *UserService) GetUserByEmailorUsername(ctx context.Context, key string, value string) (models.User, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetUserByEmailorUsername")
	defer childSpan.End()

	user, err := us.userRepository.GetUserByEmailorUsername(trContext, key, value)

	if err != nil {
		return models.User{}, err
	}

	return user, err
}

func (us *UserService) CreateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCCreateUser")
	defer childSpan.End()

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)

	if err != nil {
		us.logger.Error("failed to hash password",
			"error", err.Error(),
			"username", u.Username,
		)
		return models.UserResponse{}, common.ErrInternalError
	}

	// Replace plain text password with hashed version
	u.Password = string(hashedPassword)

	user, err := us.userRepository.CreateUser(trContext, u)

	if err != nil {
		return models.UserResponse{}, err
	}

	return user, nil
}

// communicate with another service

func (us *UserService) Login(ctx context.Context, user string, password string) (string, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCLogin")
	defer childSpan.End()

	response, err := us.GetUserByEmailorUsername(trContext, "username", user)

	if err != nil {
		childSpan.SetAttributes(attribute.String("error", err.Error()))
		return "", common.ErrWrongCredentials
	}

	// Compare hashes
	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(password))
	if err != nil {
		us.logger.Error("authentication failed, wrong password",
			"username", user,
			"error", "password mismatch",
		)
		return "", common.ErrWrongCredentials
	}

	// Password is correct, generate JWT token
	token, err := us.generateJWT(response.UserId, response.Username, response.Email)

	if err != nil {
		us.logger.Error("error generating token",
			"error", err.Error(),
			"username", user,
		)
		return "", common.ErrInternalError
	}

	return token, nil

}

// generateJWT creates a JWT token with user claims
func (us *UserService) generateJWT(userID, username, email string) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	// Use the service's JWT secret (not hardcoded!)
	tokenString, err := token.SignedString(us.jwtSecret)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
