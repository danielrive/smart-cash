package main

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/go-faker/faker/v4"
)

type User struct {
	FirstsName string `json:"firstsName" faker:"first_name"`
	LastName   string `json:"lastName" faker:"last_name"`
	Username   string `json:"username" faker:"username"`
	Email      string `json:"email" faker:"email"`
	Password   string `json:"password" faker:"password"`
	UserId     string `json:"userId"`
}

type Expense struct {
	Description string  `json:"description" faker:"sentence"`
	Name        string  `json:"name" faker:"word"`
	Amount      float64 `json:"amount" faker:"amount" `
	Category    string  `json:"category" faker:"word"`
	ExpenseId   string  `json:"expenseId"`
	UserId      string  `json:"userId"`
}

func main() {
	// define api endpoint
	//apiEndpoint := "http://api.develop.smartcash.rootkit.site"
	fmt.Println(createUser("http://api.develop.smartcash.rootkit.site"))

}

func createUser(apiEndpoint string) string {
	// create user
	userURL := fmt.Sprintf("%s/user", apiEndpoint)
	// Create body for POST request
	fakeUser := User{}
	err := faker.FakeData(&fakeUser)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(userURL)

	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, // <-- skip TLS verification
		},
	}

	// Marshal the struct into JSON
	jsonData, err := json.Marshal(fakeUser)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return ""
	}
	// Create a new POST request with the JSON body
	req, err := http.NewRequest("POST", userURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")

	// Send the request
	resp, err := httpClient.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Read and print the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}
	fmt.Println(string(respBody))
	newUser := User{}
	err = json.Unmarshal(respBody, &newUser)
	if err != nil {
		fmt.Println("error could not parse response body for user")
		return ""
	}
	return newUser.UserId
}

/*
func createExpense(apiEndpoint string, userId string) string {
	// create user
	expenseURL := fmt.Sprintf("%s/expense", apiEndpoint)

	// Create body for POST request
	fakeExpense := Expense{}
	err := faker.FakeData(&fakeExpense)
	if err != nil {
		fmt.Println(err)
	}
	// Marshal the struct into JSON
	jsonData, err := json.Marshal(fakeExpense)
	if err != nil {
		fmt.Println("Error marshaling JSON:", err)
		return ""
	}
	// Create a new POST request with the JSON body
	req, err := http.NewRequest("POST", expenseURL, bytes.NewBuffer(jsonData))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return ""
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("UserId", userId)

	// Send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return ""
	}
	defer resp.Body.Close()

	// Read and print the response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response:", err)
		return ""
	}
	newExpense := Expense{}
	err = json.Unmarshal(respBody, &newExpense)
	if err != nil {
		fmt.Println("error could not parse response body for user")
		return ""
	}
	return newExpense.ExpenseId
}
*/
