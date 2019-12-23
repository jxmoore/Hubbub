package models

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/microsoft/ApplicationInsights-Go/appinsights"
)

// NotificationHandler is an interface that knows how to build out a notification and display to the
// consumer of the message.
type NotificationHandler interface {
	Init(c *Config) error
	Notify(NotificationDetails) error
}

// NotificationDetails is an struct that holds fields used by the the a specific notifications Notify method. For example
// body is used for slack and STDOUT/Default whereas Properties is used by applicationinsights.
type NotificationDetails struct {
	body       []byte
	properties map[string]string
}

// ApplicationInsights is a struct that holds the information needed to send a customEvent via
// the Notify() method.
type ApplicationInsights struct {
	eventTitle string
	key        string
	client     appinsights.TelemetryClient
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

// Init copies the key from the config into the 'a' and creates the Application Insights Client inside of 'a'
func (a *ApplicationInsights) Init(c *Config) error {

	a.key = c.Notification.AppInsightsKey
	a.client = appinsights.NewTelemetryClient(c.Notification.AppInsightsKey)
	return nil

}

// Init loads the slack config from the *Config into 's'
// An error is returned if one or more of these values is abscent
func (s *Slack) Init(c *Config) error {

	s.Title = c.Notification.SlackTitle
	s.Icon = c.Notification.SlackIcon
	s.User = c.Notification.SlackUser

	if c.Notification.SlackWebHook != "" {
		s.WebHook = c.Notification.SlackWebHook
	}
	if c.Notification.SlackChannel != "" {
		s.Channel = c.Notification.SlackChannel
	}

	if s.WebHook == "" || s.Channel == "" {
		return fmt.Errorf("Missing slack token or channel")
	}

	return nil
}

// Notify submits the custom event in application insights
func (a ApplicationInsights) Notify(details NotificationDetails) error {

	event := appinsights.NewEventTelemetry(a.eventTitle)
	event.Properties = details.properties
	a.client.Track(event)

	return nil

}

// Notify is a method on Slack that posts the message to slack.
func (s Slack) Notify(details NotificationDetails) error {

	// TODO:
	// Add retry logic (assuming not a 40* result code but 500 etc..)

	client := &http.Client{}
	buffer := bytes.NewBuffer(details.body)
	request, err := http.NewRequest("POST", s.WebHook, buffer)
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

	if strings.ToLower(string(body)) == "ok" {
		fmt.Printf("Slack message sent\n")
	} else {
		fmt.Printf("Attempted to send message but received %v\n", string(body))
	}

	return nil
}

// Notify prints the message to STDOUT.
func (s STDOUT) Notify(details NotificationDetails) error {

	fmt.Println(string(details.body))
	return nil

}

// BuildBody is an exported function that takes a NotificationHandler interface and a PodStatusInformation Struct and uses these to build out a notificationDetails
// struct. The return value (notificationDetails) is used by all structs ({struct}.Notify()) that satisfy the NotificationHandeler interface.
func BuildBody(handler NotificationHandler, Pod PodStatusInformation) (NotificationDetails, error) {

	nDetails := NotificationDetails{}

	if s, ok := handler.(*Slack); ok {
		var err error
		nDetails.body, err = BuildSlackBody(*s, Pod)
		if err != nil {
			return nDetails, fmt.Errorf("There was an error creating the body for the slack post. %v", err.Error())
		}
		return nDetails, nil
	}

	Pod.ConvertTime()
	nDetails.body, _ = json.Marshal(Pod)

	nDetails.properties["Pod"] = Pod.PodName
	nDetails.properties["Container"] = Pod.ContainerName
	nDetails.properties["Namespace"] = Pod.Namespace
	nDetails.properties["Image"] = Pod.Image
	nDetails.properties["RunTime"] = fmt.Sprintf("%v until %v", Pod.StartedAt, Pod.FinishedAt)
	nDetails.properties["FailureReason"] = podErrorReason(Pod)
	nDetails.properties["ExitCode"] = podErrorCode(Pod)

	return nDetails, nil

}

// BuildSlackBody builds out the JSON payload that is used to post the message to slack.
// It takes a struct of PodStatusInformation and with that it creates the SlackAttachment struct and marshels 's'
// The marshalled 's' is returned to the caller.
func BuildSlackBody(s Slack, Pod PodStatusInformation) ([]byte, error) {

	color := "danger"
	errorDetails := podErrorCode(Pod)
	reason := podErrorReason(Pod)

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

	slackMsg, _ := json.Marshal(s)

	return slackMsg, nil

}
