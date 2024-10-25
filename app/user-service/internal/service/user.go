package service

import (
	"log"
	"net/http"
	"smart-cash/user-service/internal/repositories"
	"smart-cash/user-service/models"
)

type UserService struct {
	userRepository *repositories.DynamoDBUsersRepository
}

func NewUserService(userRepository *repositories.DynamoDBUsersRepository) *UserService {
	return &UserService{userRepository: userRepository}
}

// Function to generate random string
/**
func generateRandomToken(length int) (string, error) {
	// Calculate the byte length required for the given token length
	byteLength := length / 2 // Each byte encodes 2 hexadecimal characters

	// Create a byte slice to hold the random bytes
	randomBytes := make([]byte, byteLength)

	// Fill the byte slice with random bytes
	_, err := rand.Read(randomBytes)
	if err != nil {
		log.Println("error", err)
		return "", err
	}

	// Encode the random bytes to a hexadecimal string
	token := hex.EncodeToString(randomBytes)

	return token, nil
}
**/
func (us *UserService) GetUserById(userId string) (models.UserResponse, error) {

	user, err := us.userRepository.GetUserById(userId)

	if err != nil {
		log.Println("error ", err)
		return models.UserResponse{}, err
	}

	return user, nil
}

func (us *UserService) GetUserByEmailorUsername(key string, value string) (models.UserResponse, error) {
	// Find user
	user, err := us.userRepository.GetUserByEmailorUsername(key, value)

	if err != nil {
		log.Println("error ", err)
		return models.UserResponse{}, err
	}

	return user, err
}

func (us *UserService) CreateUser(u models.User) (models.UserResponse, error) {
	// generate UUID for the user

	user, err := us.userRepository.CreateUser(u)

	if err != nil {
		log.Println("error ", err)
		return models.UserResponse{}, err
	}

	return user, nil
}

// communicate with another service

func (us *UserService) ConnectOtherSVC(svc_name string, port string) error {
	baseURL := "http://" + svc_name + ":" + port + "/health"
	resp, err := http.Get(baseURL)
	if err != nil {
		log.Println("Error creating request:", err)
		return err
	}

	// Close the response body after reading
	defer resp.Body.Close()

	// Call the internal function to validate the user token
	log.Println("response from http call ", resp.StatusCode)
	return nil

}
