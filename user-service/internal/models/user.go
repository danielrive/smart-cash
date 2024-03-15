// define struct for users
package models

type User struct {
	UserId            string `json:"userId" dynamodbav:"userId"` // primary key
	FirstsName        string `json:"firstsName" dynamodbav:"firstsName"`
	LastName          string `json:"lastName" dynamodbav:"lastName"`
	Email             string `json:"email" dynamodbav:"email"` // global secondary index
	Username          string `json:"username" dynamodbav:"username"`
	BankName          string `json:"bankName" dynamodbav:"bankName"`
	BankAccountNumber int    `json:"bankAccountNumber" dynamodbav:"bankAccountNumber"`
	Active            bool   `json:"active" dynamodbav:"active"`
}
