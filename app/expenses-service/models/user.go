package models

type User struct {
	UserId   string `json:"userId" dynamodbav:"userId"` // primary key
	Username string `json:"username" dynamodbav:"username"`
	Email    string `json:"email" dynamodbav:"email"` // global secondary index
	Active   bool   `json:"active" dynamodbav:"active"`
}
