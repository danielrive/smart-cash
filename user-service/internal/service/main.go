package service

import (
	"user-service/internal/repositories"

	"user-service/internal/models"
)

type UserService struct {
	userRepository *repositories.DynamoDBUsersRepository
}

func NewUserService(userRepository *repositories.DynamoDBUsersRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

func (us *UserService) GetUserById(userId string) (models.User, error) {

	user, err := us.userRepository.GetUserById(userId)

	if err != nil {
		return models.User{}, err
	}

	return user, nil
}

func (us *UserService) GetUserByEmail(email string) (models.User, error) {
	// Find user
	user, err := us.userRepository.FindUserByEmail(email)

	if err != nil {
		return models.User{}, err
	}

	return user, err
}

func (us *UserService) CreateUser(u models.User) error {
	// search user by email

	err := us.userRepository.CreateUser(u)

	if err != nil {
		return err
	}

	return nil
}
