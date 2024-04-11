package service

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"smart-cash/user-service/internal/common"
	"smart-cash/user-service/internal/models"
	"smart-cash/user-service/internal/repositories"
)

type UserService struct {
	userRepository *repositories.DynamoDBUsersRepository
}

func NewUserService(userRepository *repositories.DynamoDBUsersRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

// Function to generate random string

func generateRandomToken(length int) (string, error) {
	// Calculate the byte length required for the given token length
	byteLength := length / 2 // Each byte encodes 2 hexadecimal characters

	// Create a byte slice to hold the random bytes
	randomBytes := make([]byte, byteLength)

	// Fill the byte slice with random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Encode the random bytes to a hexadecimal string
	token := hex.EncodeToString(randomBytes)

	return token, nil
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
	user, err := us.userRepository.GetUserByEmail(email)

	if err != nil {
		return models.User{}, err
	}

	return user, err
}

func (us *UserService) CreateUser(u models.User) error {
	// generate UUID for the user

	err := us.userRepository.CreateUser(u)

	if err != nil {
		return err
	}

	return nil
}

/// login service

// funct that return user id

func (us *UserService) Login(u models.User) (string, string, error) {
	// Find user
	user, err := us.userRepository.GetUserByEmail(u.Email)
	if err != nil {
		return "", "", err
	}

	fmt.Println("enter to valdiate pass")
	if user.Password == u.Password {
		token, err := generateRandomToken(32)
		user.Token = token
		if err != nil {
			return "", "", err
		} else {
			log.Println("User validated ")
			us.userRepository.UpdateUser(user)
			return user.UserId, token, nil
		}
	}
	log.Println("error", common.ErrWrongCredentials)
	return "", "", common.ErrWrongCredentials
}
