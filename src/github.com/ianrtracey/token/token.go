package token

import (
	"io/ioutil"
)

type SlackToken struct {
	Token string `json:"api_token"`
}

func GetToken() ([]byte, error) {
	return ioutil.ReadFile("./token.json")
}
