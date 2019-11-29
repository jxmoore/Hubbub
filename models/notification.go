package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// NotificationHandler is an interface that knows how to build out a notification and display to the
// consumer of the message.
type NotificationHandler interface {
	Init(c *Config) error
	BuildBody(Pod PodStatusInformation) (string, error)
	Notify(msg string) error
}

// Slack is a struct that stores the Slack config, and the post body structs (SlackAttachments[SlackFields])
type Slack struct {
	WebHook    string             `json:"-"`
	Title      string             `json:"-"`
	Channel    string             `json:"channel"`
	User       string             `json:"username"`
	Icon       string             `json:"icon_url"`
	Attachment []SlackAttachments `json:"attachments"`
}

// SlackAttachments is a struct that represents the attachment portion of a slack payload.
type SlackAttachments struct {
	Fallback string        `json:"fallback"`
	Color    string        `json:"color"`
	Title    string        `json:"title"`
	Field    []SlackFields `json:"fields"`
}

// SlackFields is a struct that represents the Fields portion of a slack payload.
type SlackFields struct {
	Value string `json:"value"`
}

// STDOUT is a small struct used to hold a json payload thats printed to the screen.
type STDOUT struct {
	Body string
}

// Init servers no purpose other to satisfy the interface
func (s *STDOUT) Init(c *Config) error {
	return nil
}

// Init loads the slack config from the *Config into 's'
// An error is returned if one or more of these values is abscent
func (s *Slack) Init(c *Config) error {

	s.Title = c.Slack.Title
	s.Icon = c.Slack.Icon
	s.User = c.Slack.User

	if c.Slack.WebHook != "" {
		s.WebHook = c.Slack.WebHook
	}
	if c.Slack.Channel != "" {
		s.Channel = c.Slack.Channel
	}

	if s.WebHook == "" || s.Channel == "" {
		return fmt.Errorf("Missing slack token or channel")
	}

	return nil
}

// BuildBody builds out the JSON payload that is used to post the message to slack.
// It takes a struct of PodStatusInformation and with that it creates the SlackAttachment struct and marshels 's'
// The marshalled 's' is returned to the caller.
func (s Slack) BuildBody(Pod PodStatusInformation) (string, error) {

	var reason string
	color := "danger"
	errorDetails := strconv.Itoa(Pod.ExitCode)
	errInfo := Pod.ExitCodeLookup()

	if errInfo != "" {
		errorDetails = fmt.Sprintf("Error code : %v `%v`", errorDetails, errInfo)
	} else {
		errorDetails = fmt.Sprintf("Error code : %v\n", errorDetails)
	}

	if Pod.Reason != "" && Pod.Message != "" {
		reason = fmt.Sprintf("Failure reason received : `%v - %v`", Pod.Reason, Pod.Message)
	} else if Pod.Message != "" {
		reason = fmt.Sprintf("Failure reason received : `%v`", Pod.Message)
	} else if Pod.Reason != "" {
		reason = fmt.Sprintf("Failure reason received : `%v`", Pod.Reason)
	} else {
		reason = "Unable to determine the reason for the failure."
	}

	Pod.ConvertTime()

	// time.Format returns a string, to get out of having another field in the struct we format it here in line.
	msg := fmt.Sprintf("The pod : *%v* has encountered an error.\n\nThe container is : *%v*\nWhich is running image : *%v*.\nThe error information is below.\n\n\n"+
		"> %v\n> %v\n> The pod ran from : *%v until %v*", Pod.PodName, Pod.ContainerName, Pod.Image, reason, errorDetails,
		Pod.StartedAt.Format(time.Stamp), Pod.FinishedAt.Format(time.Stamp))

	s.Attachment = []SlackAttachments{
		SlackAttachments{
			Fallback: msg,
			Color:    color,
			Title:    s.Title,
			Field:    []SlackFields{SlackFields{Value: msg}},
		},
	}

	slackMsg, err := json.Marshal(s)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}

	return string(slackMsg), nil

}

// BuildBody builds out the JSON payload that is used written to the screen when using STDOUT.
func (s STDOUT) BuildBody(Pod PodStatusInformation) (string, error) {

	msg, err := json.Marshal(Pod)
	if err != nil {
		return "", fmt.Errorf(err.Error())
	}
	return string(msg), nil

}

// Notify is a method on Slack that posts the message to slack.
func (s Slack) Notify(msg string) error {

	// TODO:
	// Add retry logic (assuming not a 40* result code but 500 etc..)
	// Check response to ensure we receive back the 'ok' msg that slack replies with.

	client := &http.Client{}
	buffer := []byte(msg)

	request, err := http.NewRequest("POST", s.WebHook, bytes.NewBuffer(buffer))
	if err != nil {
		return errors.New(err.Error())
	}

	request.Header.Add("Content-Type", "application/json")
	response, err := client.Do(request)
	if err != nil {
		return errors.New(err.Error())
	}

	defer response.Body.Close()

	// Output is just dropped as is, not unmarsheled
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return errors.New(err.Error())
	}

	fmt.Printf("Slack message sent \n%v", string(body))

	return nil
}

// Notify prints the message to STDOUT.
func (s STDOUT) Notify(msg string) error {

	fmt.Println(msg)
	return nil

}
