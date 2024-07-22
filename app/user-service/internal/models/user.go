// define struct for users
package models

type User struct {
	UserId     string `json:"userId" dynamodbav:"userId"` // primary key
	FirstsName string `json:"firstsName" dynamodbav:"firstsName"`
	LastName   string `json:"lastName" dynamodbav:"lastName"`
	Username   string `json:"username" dynamodbav:"username"`
	Email      string `json:"email" dynamodbav:"email"` // global secondary index
	Password   string `json:"password" dynamodbav:"password"`
	Active     bool   `json:"active" dynamodbav:"active"`
	Token      string `json:"token" dynamodbav:"token"`
}

type UserResponse struct {
	UserId   string `json:"userId" dynamodbav:"userId"` // primary key
	Username string `json:"username" dynamodbav:"username"`
	Email    string `json:"email" dynamodbav:"email"` // global secondary index
	Active   bool   `json:"active" dynamodbav:"active"`
}
