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

	us.logger.Debug("getting user by id",
		slog.String("user_id", userId),
		slog.String("component", "service"),
	)

	user, err := us.userRepository.GetUserById(trContext, userId)
	if err != nil {
		us.logger.Error("error getting user by id",
			slog.String("user_id", userId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.UserResponse{}, err
	}

	us.logger.Debug("user retrieved successfully",
		slog.String("user_id", userId),
		slog.String("component", "service"),
	)

	return user, nil
}

func (us *UserService) GetUserByEmailorUsername(ctx context.Context, key string, value string) (models.User, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCGetUserByEmailorUsername")
	defer childSpan.End()

	us.logger.Debug("getting user by email or username",
		slog.String("key", key),
		slog.String("value", value),
		slog.String("component", "service"),
	)

	user, err := us.userRepository.GetUserByEmailorUsername(trContext, key, value)
	if err != nil {
		us.logger.Debug("user not found by email or username",
			slog.String("key", key),
			slog.String("value", value),
			slog.String("component", "service"),
		)
		return models.User{}, err
	}

	us.logger.Debug("user found by email or username",
		slog.String("user_id", user.UserId),
		slog.String("key", key),
		slog.String("component", "service"),
	)

	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCCreateUser")
	defer childSpan.End()

	us.logger.Info("creating new user",
		slog.String("username", u.Username),
		slog.String("email", u.Email),
		slog.String("component", "service"),
	)

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		us.logger.Error("failed to hash password",
			slog.String("error", err.Error()),
			slog.String("username", u.Username),
			slog.String("component", "service"),
		)
		return models.UserResponse{}, common.ErrInternalError
	}

	// Replace plain text password with hashed version
	u.Password = string(hashedPassword)

	user, err := us.userRepository.CreateUser(trContext, u)
	if err != nil {
		us.logger.Error("error creating user in repository",
			slog.String("username", u.Username),
			slog.String("email", u.Email),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.UserResponse{}, err
	}

	us.logger.Info("user created successfully",
		slog.String("user_id", user.UserId),
		slog.String("username", user.Username),
		slog.String("component", "service"),
	)

	return user, nil
}

// communicate with another service

func (us *UserService) Login(ctx context.Context, user string, password string) (string, error) {
	tr := otel.Tracer(common.ServiceName)
	trContext, childSpan := tr.Start(ctx, "SVCLogin")
	defer childSpan.End()

	us.logger.Debug("attempting login",
		slog.String("username", user),
		slog.String("component", "service"),
	)

	response, err := us.GetUserByEmailorUsername(trContext, "username", user)
	if err != nil {
		childSpan.SetAttributes(attribute.String("error", err.Error()))
		us.logger.Warn("login failed - user not found",
			slog.String("username", user),
			slog.String("component", "service"),
		)
		return "", common.ErrWrongCredentials
	}

	// Compare hashes
	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(password))
	if err != nil {
		us.logger.Warn("login failed - password mismatch",
			slog.String("username", user),
			slog.String("user_id", response.UserId),
			slog.String("component", "service"),
		)
		return "", common.ErrWrongCredentials
	}

	// Password is correct, generate JWT token
	token, err := us.generateJWT(response.UserId, response.Username, response.Email)
	if err != nil {
		us.logger.Error("error generating JWT token",
			slog.String("error", err.Error()),
			slog.String("username", user),
			slog.String("user_id", response.UserId),
			slog.String("component", "service"),
		)
		return "", common.ErrInternalError
	}

	us.logger.Info("login successful",
		slog.String("username", user),
		slog.String("user_id", response.UserId),
		slog.String("component", "service"),
	)

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
