package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/ianrtracey/ballot"
	"github.com/ianrtracey/token"
	"github.com/looplab/fsm"
	"github.com/nlopes/slack"
)

type Message struct {
	Text string
}

func Map(vs []string, f func(string) string) []string {
	vsm := make([]string, len(vs))
	for i, v := range vs {
		vsm[i] = f(v)
	}
	return vsm
}

func HandleParliamentMessage(ev *slack.MessageEvent, rtm *slack.RTM, trial *fsm.FSM, ballot *ballot.Ballot, slack_api *slack.Client) {
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
		trial.Event("submit_topic")
		rtm.SendMessage(rtm.NewOutgoingMessage("Cool, tell me some of the voting options:", ev.Channel))
		return

	case "waiting_on_items":
		// need to add check to verify that the speaker of the house actually submitted at least two items
		if strings.Contains(ev.Text, "done") {
			trial.Event("complete_items")
			rtm.SendMessage(rtm.NewOutgoingMessage("Awesome! Commence the voting! DM'ing the channel...", ev.Channel))
			channel_info, err := slack_api.GetChannelInfo(ev.Channel)
			if err != nil {
				panic(fmt.Sprintf("Something went wrong getting the channel infomation for %v\n", ev.Channel))
			}
			fmt.Println(channel_info)
			rtm.SendMessage(rtm.NewOutgoingMessage("Awesome! Commence the voting! DM'ing the channel...", "U02NWHMNG"))
			ok, already_open, channel_id, err := slack_api.OpenIMChannel("U02NWHMNG")
			if err != nil {
				panic("Something went wrong!")
			}
			fmt.Println("channel %v\n %v\n %v\n", channel_id, ok, already_open)
			// This buils the options string
			message := fmt.Sprintf(strings.Join(ballot.Items, "\n"))
			rtm.SendMessage(rtm.NewOutgoingMessage(message, channel_id))

			trial.Event("open_polling")
			return
		}
		ballot.AddItem(ev.Text)
		fmt.Println(ballot.Items)

	case "polling":
		fmt.Println(ev.Text)

	default:
		panic("undhandled case!")
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
			{Name: "complete_items", Src: []string{"waiting_on_items"}, Dst: "preparing_ballot"},
			{Name: "open_polling", Src: []string{"preparing_ballot"}, Dst: "polling"},
		},
		fsm.Callbacks{},
	)
	ballot := &ballot.Ballot{}

	logger := log.New(os.Stdout, "slack-bot:", log.Lshortfile|log.LstdFlags)
	// need to set bot_handle to be dynamic as the id could change depending on the instance
	bot_handle := "@U1QAH7PRD"
	slack.SetLogger(logger)
	var slack_token token.SlackToken
	token_file, err := token.GetToken()

	if err != nil {
		fmt.Printf("File Error: %s\n", err)
		os.Exit(1)
	}

	json.Unmarshal(token_file, &slack_token)
	slack_api := slack.New(slack_token.Token)
	slack_api.SetDebug(true)

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
				fmt.Println("Connection counter:", ev.ConnectionCount)

			case *slack.MessageEvent:
				fmt.Printf("Message: %v\n", ev)
				if strings.Contains(ev.Text, bot_handle) {
					HandleParliamentMessage(ev, rtm, trial, ballot, slack_api)
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
