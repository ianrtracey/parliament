package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/looplab/fsm"
	"github.com/nlopes/slack"
)

type Message struct {
	Text string
}

type SlackToken struct {
	Token string `json:"api_token"`
}

type Trial struct {
	Topic string
	Items []string
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

func HandleParliamentMessage(ev *slack.MessageEvent, rtm *slack.RTM, trial *fsm.FSM) {
	fmt.Println(trial.Current())
	switch trial.Current() {
	case "inactive":
		rtm.SendMessage(rtm.NewOutgoingMessage("Here ye, here ye! Should we hold a trial?", ev.Channel))
		trial.Event("start")

	case "awaiting_confirmation_of_trial":
		if strings.Contains(ev.Text, "yes") {
			rtm.SendMessage(rtm.NewOutgoingMessage("Cool, let's get started :firework:", ev.Channel))
			trial.Event("confirm_trial")
			rtm.SendMessage(rtm.NewOutgoingMessage("What is the purpose this trial?", ev.Channel))
			return
		}

		if strings.Contains(ev.Text, "no") {
			rtm.SendMessage(rtm.NewOutgoingMessage("Alright, see you next time :cry:", ev.Channel))
			trial.Event("decline_trial")
			return
		}
		rtm.SendMessage(rtm.NewOutgoingMessage("I didn't quite catch that", ev.Channel))

	case "waiting_on_topic":
		rtm.SendMessage(rtm.NewOutgoingMessage("Cool, tell me some of the voting options:", ev.Channel))
		trial.Event("submit_topic")
		return

	default:
		panic("unhandled case!")
	}
}

func main() {

	trial := fsm.NewFSM(
		"inactive",
		fsm.Events{
			{Name: "start", Src: []string{"inactive"}, Dst: "awaiting_confirmation_of_trial"},
			{Name: "confirm_trial", Src: []string{"awaiting_confirmation_of_trial"}, Dst: "waiting_on_topic"},
			{Name: "decline_trial", Src: []string{"awaiting_confirmation_of_trial"}, Dst: "inactive"},
			{Name: "submit_topic", Src: []string{"waiting_on_topic"}, Dst: "waiting_on_items"},
		},
		fsm.Callbacks{},
	)

	logger := log.New(os.Stdout, "slack-bot:", log.Lshortfile|log.LstdFlags)
	// need to set bot_handle to be dynamic as the id could change depending on the instance
	bot_handle := "@U1QAH7PRD"
	slack.SetLogger(logger)
	var slack_token SlackToken
	token_file, err := getToken()

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
		case msg := <-rtm.IncomingEvents:
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
					HandleParliamentMessage(ev, rtm, trial)
				}
				fmt.Printf("Trial: %v\n", trial)

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
