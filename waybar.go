package main

import (
	"encoding/json"
	"fmt"
)

type waybarOutput struct {
	Text    string `json:"text"`
	Tooltip string `json:"tooltip,omitempty"`
	Class   string `json:"class,omitempty"`
}

func printWaybarJSON(out waybarOutput) {
	data, _ := json.Marshal(out)
	fmt.Println(string(data))
}

func printWaybarError(msg string) {
	printWaybarJSON(waybarOutput{
		Text:    "󰀨",
		Tooltip: msg,
		Class:   "linear-error",
	})
}

func runWaybar(client *Client) {
	count, err := client.FetchUnreadCount()
	if err != nil {
		printWaybarError(err.Error())
		return
	}
	if count <= 0 {
		printWaybarJSON(waybarOutput{Text: ""})
		return
	}
	printWaybarJSON(waybarOutput{
		Text:    fmt.Sprintf("󰔖 %d", count),
		Tooltip: fmt.Sprintf("%d unread Linear notifications", count),
		Class:   "linear-unread",
	})
}
