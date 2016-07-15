package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"github.com/nlopes/slack"
	"io/ioutil"
	"os"
)

type Message struct {
	Text string
}

type SlackToken struct {
	Token string `json:"api_token"`
}

func heartBeatHandler(w http.ResponseWriter, r *http.Request) {
	m := Message{"OK"}
	resp, err := json.Marshal(m)

	if err != nil {
		panic(err)
	}

	w.Write(resp)
}




func main() {
	port := ":8080"
	token_file, err  := ioutil.ReadFile("./token.json")
	if err != nil {
		fmt.Printf("File Error: %s\n", err)
		os.Exit(1)
	}
	var slack_token SlackToken
	json.Unmarshal(token_file, &slack_token)
	fmt.Printf("result %s\n", slack_token.Token)
	slack_api := slack.New(slack_token.Token)
	slack_api.SetDebug(true)
	groups, err := slack_api.GetChannels(false)
	if err != nil {
		fmt.Printf("%s\n", err)
		return
	}
	for _, group:= range groups {
		fmt.Printf("yo yo yo %s", group.ID)
	}
	http.HandleFunc("/heartbeat", heartBeatHandler)
	fmt.Printf("Serving on port %s\n", port)
	http.ListenAndServe(port, nil)
}
