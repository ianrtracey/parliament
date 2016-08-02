package trial

import (
	"github.com/looplab/fsm"
)

func BuildNewTrial() {
	return fsm.NewFSM(
		"inactive",
		fsm.Events{
			{Name: "start", Src: []string{"inactive"}, Dst: "awaiting_confirmation_of_trial"},
			{Name: "confirm_trial", Src: []string{"awaiting_confirmation_of_trial"}, Dst: "waiting_on_topic"},
			{Name: "decline_trial", Src: []string{"awaiting_confirmation_of_trial"}, Dst: "inactive"},
			{Name: "submit_topic", Src: []string{"waiting_on_topic"}, Dst: "waiting_on_items"},
			{Name: "complete_items", Src: []string{"waiting_on_items"}, Dst: "preparing_ballot"},
		},
		fsm.Callbacks{},
	)
}
