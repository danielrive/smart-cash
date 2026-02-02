package service

import (
	"context"
	"log/slog"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/models"
	"smart-cash/utils"
	"smart-cash/utils/logging"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.opentelemetry.io/otel/attribute"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	userRepository *repositories.DynamoDBUsersRepository
	logger         *slog.Logger
	jwtSecret      []byte
}

// loggerWithTrace returns a logger with trace context if available in the context
func (us *UserService) loggerWithTrace(ctx context.Context) *slog.Logger {
	return logging.LoggerWithTraceContext(ctx, us.logger)
}

type claims = struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
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
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCGetUserById", "service",
		attribute.String("user.id", userId),
	)
	defer endSpan()

	us.loggerWithTrace(ctx).Debug("getting user by id",
		slog.String("user_id", userId),
		slog.String("component", "service"),
	)

	user, err := us.userRepository.GetUserById(ctx, userId)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		us.loggerWithTrace(ctx).Error("error getting user by id",
			slog.String("user_id", userId),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.UserResponse{}, err
	}

	us.loggerWithTrace(ctx).Debug("user retrieved successfully",
		slog.String("user_id", userId),
		slog.String("component", "service"),
	)

	return user, nil
}

func (us *UserService) GetUserByEmailorUsername(ctx context.Context, key string, value string) (models.User, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCGetUserByEmailorUsername", "service",
		attribute.String("query.key", key),
		attribute.String("query.value", value),
	)
	defer endSpan()

	us.loggerWithTrace(ctx).Info("getting user by email or username",
		slog.String("key", key),
		slog.String("value", value),
		slog.String("component", "service"),
	)

	user, err := us.userRepository.GetUserByEmailorUsername(ctx, key, value)
	if err != nil {
		us.loggerWithTrace(ctx).Warn("user not found by email or username",
			slog.String("key", key),
			slog.String("value", value),
			slog.String("component", "service"),
		)
		return models.User{}, err
	}

	us.loggerWithTrace(ctx).Info("user found by email or username",
		slog.String("user_id", user.UserId),
		slog.String("key", key),
		slog.String("component", "service"),
	)

	return user, nil
}

func (us *UserService) CreateUser(ctx context.Context, u models.User) (models.UserResponse, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCCreateUser", "service",
		attribute.String("user.username", u.Username),
		attribute.String("user.email", u.Email),
	)
	defer endSpan()

	us.loggerWithTrace(ctx).Info("creating new user",
		slog.String("username", u.Username),
		slog.String("email", u.Email),
		slog.String("component", "service"),
	)

	// Hash the password before storing
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		us.loggerWithTrace(ctx).Error("failed to hash password",
			slog.String("error", err.Error()),
			slog.String("username", u.Username),
			slog.String("component", "service"),
		)
		return models.UserResponse{}, common.ErrInternalError
	}

	// Replace plain text password with hashed version
	u.Password = string(hashedPassword)

	user, err := us.userRepository.CreateUser(ctx, u)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		us.loggerWithTrace(ctx).Error("error creating user in repository",
			slog.String("username", u.Username),
			slog.String("email", u.Email),
			slog.String("error", err.Error()),
			slog.String("component", "service"),
		)
		return models.UserResponse{}, err
	}

	us.loggerWithTrace(ctx).Info("user created successfully",
		slog.String("user_id", user.UserId),
		slog.String("username", user.Username),
		slog.String("component", "service"),
	)

	return user, nil
}

// communicate with another service

func (us *UserService) Login(ctx context.Context, user string, password string) (string, error) {
	ctx, endSpan := utils.StartSpanWithComponent(ctx, common.ServiceName, "SVCLogin", "service",
		attribute.String("user.username", user),
	)
	defer endSpan()

	us.loggerWithTrace(ctx).Debug("attempting login",
		slog.String("username", user),
		slog.String("component", "service"),
	)

	response, err := us.GetUserByEmailorUsername(ctx, "username", user)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		us.loggerWithTrace(ctx).Warn("login failed - user not found",
			slog.String("username", user),
			slog.String("component", "service"),
		)
		return "", common.ErrWrongCredentials
	}

	// Check if user is active
	us.loggerWithTrace(ctx).Info("checking if user is active",
		slog.String("username", user),
		slog.String("user_id", response.UserId),
		slog.Bool("active", response.Active),
		slog.String("component", "service"),
	)
	if !response.Active {
		us.loggerWithTrace(ctx).Warn("login failed - user account is inactive",
			slog.String("username", user),
			slog.String("user_id", response.UserId),
			slog.String("component", "service"),
		)
		return "", common.ErrUserInactive
	}

	// Compare hashes
	err = bcrypt.CompareHashAndPassword([]byte(response.Password), []byte(password))
	if err != nil {
		us.loggerWithTrace(ctx).Warn("login failed - password mismatch",
			slog.String("username", user),
			slog.String("user_id", response.UserId),
			slog.String("component", "service"),
		)
		return "", common.ErrWrongCredentials
	}

	// Password is correct, generate JWT token
	token, err := us.generateJWT(response.UserId, response.Username, response.Email, response.Active)
	if err != nil {
		utils.RecordSpanError(ctx, err)
		us.loggerWithTrace(ctx).Error("error generating JWT token",
			slog.String("error", err.Error()),
			slog.String("username", user),
			slog.String("user_id", response.UserId),
			slog.String("component", "service"),
		)
		return "", common.ErrInternalError
	}

	us.loggerWithTrace(ctx).Info("login successful",
		slog.String("username", user),
		slog.String("user_id", response.UserId),
		slog.String("component", "service"),
	)

	return token, nil

}

// generateJWT creates a JWT token with user claims
func (us *UserService) generateJWT(userID, username, email string, active bool) (string, error) {
	expirationTime := time.Now().Add(1 * time.Hour)
	claims := &claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		Active:   active,
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
