package goclio

import "fmt"

type Account struct {
	client *Client
}

type LoginRequest struct {
	Email      string `json:"email"`
	Password   string `json:"password"`
	RememberMe bool   `json:"remember_me"`
}

type LoginResponse struct {
	Redirect string `json:"redirect"`
}

func (c *Account) Login(email string, password string) (LoginResponse, error) {
	res := new(LoginResponse)

	reqBody := LoginRequest{
		Email:      email,
		Password:   password,
		RememberMe: true,
	}

	endpoint := fmt.Sprintf("session.json")

	err := c.client.request("POST", endpoint, reqBody, res)
	return *res, err
}
