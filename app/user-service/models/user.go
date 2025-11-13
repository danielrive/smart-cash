// define struct for users
package models

type User struct {
	UserId    string `json:"userId" dynamodbav:"userId"` // primary key
	FirstName string `json:"firstName" dynamodbav:"firstName"`
	LastName  string `json:"lastName" dynamodbav:"lastName"`
	Username  string `json:"username" dynamodbav:"username"`
	Email     string `json:"email" dynamodbav:"email"` // global secondary index
	Password  string `json:"password" dynamodbav:"password"`
	Active    bool   `json:"active" dynamodbav:"active"`
	Token     string `json:"token" dynamodbav:"token"`
}

type UserResponse struct {
	UserId   string `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Active   bool   `json:"active"`
}

type LoginRequest struct {
	Username string `json:"username" dynamodbav:"username"`
	Password string `json:"password" dynamodbav:"password"`
}
