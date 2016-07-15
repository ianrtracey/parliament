package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"io/ioutil"
	"os"
	"log"
	"strings"

	"github.com/nlopes/slack"
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

func getToken() ([]byte, error) {
	return ioutil.ReadFile("./token.json")
}



func main() {
	logger := log.New(os.Stdout, "slack-bot:", log.Lshortfile|log.LstdFlags)
	bot_handle := "@U1QAH7PRD"
	slack.SetLogger(logger)
	var slack_token SlackToken
	token_file, err  := getToken()
	
	if err != nil {
		fmt.Printf("File Error: %s\n", err)
		os.Exit(1)
	}

	json.Unmarshal(token_file, &slack_token)
	slack_api := slack.New(slack_token.Token)
	slack_api.SetDebug(true)

	http.HandleFunc("/heartbeat", heartBeatHandler)

	rtm := slack_api.NewRTM()
	go rtm.ManageConnection()

Loop:
	for {
		select {
		case msg := <- rtm.IncomingEvents:
				fmt.Print("Event Received: ")
				switch ev := msg.Data.(type) {
					case *slack.HelloEvent:
						// Ignore

					case *slack.ConnectedEvent:
						fmt.Println("Infos:", ev.Info)
						fmt.Println("COnnection counter:", ev.ConnectionCount)
						rtm.SendMessage(rtm.NewOutgoingMessage("Hello World", "#parliament-test"))

					case *slack.MessageEvent:
						fmt.Printf("Message: %v\n", ev)
						if strings.Contains(ev.Text, bot_handle) {
							rtm.SendMessage(rtm.NewOutgoingMessage("Here ye, here ye! Should we hold a trial?", ev.Channel))
						}


					case *slack.PresenceChangeEvent:
						fmt.Printf("Presence Change %v\n", ev)

					case *slack.LatencyReport:
						fmt.Printf("Current latency: %v\n", ev.Value)

					case *slack.RTMError:
						fmt.Printf("Error: %s\n", ev.Error())

					case *slack.InvalidAuthEvent:
						fmt.Printf("Invalid credentials")
						break Loop

					default:
						// All events ignored
				}	
		}
	}
}
