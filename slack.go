package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net"
	"net/http"
	"time"

	"github.com/spf13/viper"
)

type SlackMessage struct {
	Channel     string            `json:"channel,omitempty"`
	Text        string            `json:"text,omitempty"`
	Attachments []SlackAttachment `json:"attachments,omitempty"`
	User        string            `json:"user,omitempty"`
	Ts          string            `json:"ts,omitempty"`
	ThreadTs    string            `json:"thread_ts,omitempty"`
	Update      bool              `json:"-"`
	Ephemeral   bool              `json:"-"`
}

type SlackAttachment struct {
	Title      string        `json:"title,omitempty"`
	Fallback   string        `json:"fallback,omitempty"`
	Fields     []SlackField  `json:"fields,omitempty"`
	CallbackID string        `json:"callback_id,omitempty"`
	Color      string        `json:"color,omitempty"`
	Actions    []SlackAction `json:"actions,omitempty"`
}

type SlackField struct {
	Title string `json:"title,omitempty"`
	Value string `json:"value,omitempty"`
	Short bool   `json:"short,omitempty"`
}

type SlackAction struct {
	Name    string            `json:"name,omitempty"`
	Text    string            `json:"text,omitempty"`
	Type    string            `json:"type,omitempty"`
	Value   string            `json:"value,omitempty"`
	Confirm map[string]string `json:"confirm,omitempty"`
	Style   string            `json:"style,omitempty"`
	URL     string            `json:"url,omitempty"`
}

// SlackInteraction is a struct that descibes what we
// would receive on our intereactions endpoint from slack
type SlackInteraction struct {
	Type        string            `json:"type,omitempty"`
	Actions     []SlackAction     `json:"actions,omitempty"`
	CallbackID  string            `json:"callback_id,omitempty"`
	Team        map[string]string `json:"team,omitempty"`
	Channel     map[string]string `json:"channel,omitempty"`
	User        map[string]string `json:"user,omitempty"`
	MessageTs   string            `json:"message_ts,omitempty"`
	OrigMessage SlackMessage      `json:"original_message,omitempty"`
}

// sendSlack() sends a slack message.
// It expects that viper can find "slackToken".
func sendSlack(message SlackMessage) (response []byte, err error) {

	netTransport := &http.Transport{
		Dial: (&net.Dialer{
			Timeout: 5 * time.Second,
		}).Dial,
		TLSHandshakeTimeout: 5 * time.Second,
	}

	netClient := &http.Client{
		Timeout:   time.Second * 10,
		Transport: netTransport,
	}

	slackMessage, err := json.MarshalIndent(message, "", "  ")
	if err != nil {
		return
	}

	slackToken := viper.GetString("slackToken")
	slackBytes := bytes.NewBuffer(slackMessage)

	endpoint := "https://slack.com/api/chat.postMessage"
	if message.Update && message.Ts != "" {
		endpoint = "https://slack.com/api/chat.update"
	} else if message.Ephemeral && message.User != "" {
		endpoint = "https://slack.com/api/chat.postEphemeral"
	}

	req, err := http.NewRequest("POST", endpoint, slackBytes)
	req.Header.Add("Authorization", "Bearer "+slackToken)
	req.Header.Add("Content-Type", "application/json")
	if err != nil {
		return
	}

	resp, err := netClient.Do(req)
	if err != nil {
		return
	}

	response, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}

	type SlackResponse struct {
		Ok    bool
		Error string
	}

	var slackR SlackResponse
	err = json.Unmarshal(response, &slackR)
	if err != nil {
		return
	}

	if !slackR.Ok {
		err = errors.New(slackR.Error)
	}

	return
}
