package models

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"
)

// handler is a package wide NotificationHandler used in all of the model tests as a base.
var handler NotificationHandler

// TestNotificationInit tests the init(*config) function for the notification hadler.
// Only the *slack implementation does any true 'work' and thats assignment of values from the *config to a *slack and
// verifying that fields are not nil. So this test only verifies the return values (error) are what we expect.
func TestNotificationInit(t *testing.T) {

	testSuite := map[string]struct {
		notificationType string
		expectedResponse string
		webhook          string
		channel          string
	}{
		"(s *Slack) Init() will return nil": {
			notificationType: "slack",
			channel:          "#broTalk",
			webhook:          "https://github.com/jxmoore/Hubbub/tree/develop",
		},
		"(s *STDOUT) Init() will return nil": {
			notificationType: "stdout",
		},
		"(s *Slack) Init() will throw an error due to missing fields": {
			notificationType: "slack",
			expectedResponse: "Missing slack token or channel",
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		notification := handler
		notification = new(STDOUT)

		if testCase.notificationType == "slack" {
			notification = new(Slack)
		}

		fakeConf := configFile // from config_test.go
		fakeConf.Slack.WebHook = testCase.webhook
		fakeConf.Slack.Channel = testCase.channel

		if err := notification.Init(&fakeConf); err != nil {
			if testCase.expectedResponse != err.Error() {
				t.Errorf("Expected an error code of %v and received %v", testCase.expectedResponse, err)
			} else {
				t.Logf("Received the correct error response from Init()")
			}
		}
	}
}

// TestBuildBody will call the BuildBody function on both STDOUT and Slack using a customized PodStatusInformation{}.
// The response from these is re-marshelled and we verify the expected fields are what we passed in and expect.
func TestBuildBody(t *testing.T) {

	// Theses are dummy values that will be present on the returned json body
	// when calling (s *Slack) BuildBody
	configFile.Slack.User = "hubbub"
	configFile.Slack.Icon = "https://www.sampalm.com/images/me.jpg"
	configFile.Slack.Channel = "Testing"
	configFile.Slack.WebHook = "https://www.sampalm.com/images/me.jpg"

	testSuite := map[string]struct {
		notificationType string
		exitCode         int
		podReason        string
		podMessage       string
	}{
		"The unmarshelled json from '(s Slack) BuildBody' should contain the expected strings": {
			notificationType: "slack",
			exitCode:         130,
			podReason:        "missing config",
			podMessage:       "file not found",
		},
		"The unmarshelled json from '(s Slack) BuildBody' should contain the expected strings - NIL podMessage": {
			notificationType: "slack",
			exitCode:         130,
			podReason:        "missing config",
			podMessage:       "",
		},
		"The unmarshelled json from '(s Slack) BuildBody' should contain the expected strings - NIL podReason": {
			notificationType: "slack",
			exitCode:         130,
			podReason:        "",
			podMessage:       "file not found",
		},
		"The unmarshelled json from '(s Slack) BuildBody' should contain the expected strings - NIL podReason & Message": {
			notificationType: "slack",
			exitCode:         130,
			podReason:        "",
			podMessage:       "",
		},
		"The unmarshelled json from '(s Slack) BuildBody' should contain the expected strings - Unknown exit code": {
			notificationType: "slack",
			exitCode:         72,
			podReason:        "missing config",
			podMessage:       "file not found",
		},
		"The unmarshelled json from '(s STDOUT) BuildBody' should match the pod definition provided": {
			notificationType: "stdout",
			exitCode:         139,
			podReason:        "missing config",
			podMessage:       "file not found",
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		p := Pod // from kube_test.go
		p.Namespace = "default"
		p.Image = "hubbub"
		p.PodName = "hubbubTestPod-" + testCase.notificationType
		p.ContainerName = "hubbubTestContainer"
		p.ExitCode = testCase.exitCode
		p.Reason = testCase.podReason
		p.Message = testCase.podMessage
		p.Seen = time.Now()
		bodyHandler := handler

		if testCase.notificationType == "stdout" {
			bodyHandler = new(STDOUT)
		} else {
			bodyHandler = new(Slack)
		}

		c := configFile
		c.Slack.WebHook = "google.com"
		c.Slack.Title = "Oh no!"

		bodyHandler.Init(&c)
		p.ConvertTime()
		msgInBytes, _ := bodyHandler.BuildBody(p)

		if testCase.notificationType == "stdout" {

			pCheck := PodStatusInformation{}
			json.Unmarshal(msgInBytes, &pCheck)

			// avoiding deepequal here due to issues with timestamps
			if p.Namespace == pCheck.Namespace && p.ContainerName == pCheck.ContainerName && p.Reason == pCheck.Reason &&
				p.ExitCode == pCheck.ExitCode && p.Message == pCheck.Message {
				t.Logf("Unmarshalled byte array from BuildBody matches the provided PodStatusInformation struct\n")
			} else {
				t.Errorf("The byte array returned from (s STDOUT) BuildBody did not match the provided PodStatusInformation struct!\n%v != %v", p, pCheck)
			}

		} else {

			errorDetails := strconv.Itoa(p.ExitCode)
			podMsg := fmt.Sprintf("The pod : *%v* has encountered an error.\n\nThe container is : *%v*\nWhich is running image : *%v*.\n",
				p.PodName, p.ContainerName, p.Image)

			slackBody := Slack{}
			json.Unmarshal(msgInBytes, &slackBody)
			msg := slackBody.Attachment[0].Fallback

			if slackBody.Attachment[0].Color != "danger" {
				t.Errorf("Expected slack.attachment.color == danger but received %v", slackBody.Attachment[0].Color)
			}

			if slackBody.Attachment[0].Title != c.Slack.Title {
				t.Errorf("Expected slack.attachment.title to match the title supplied in the handler init but received %v", slackBody.Attachment[0].Title)
			}

			if !strings.Contains(msg, errorDetails) {
				t.Errorf("Expected slack.attachment.fallback to contain the error details but the string %v was not found", errorDetails)
			}

			if !strings.Contains(msg, podMsg) {
				t.Errorf("Expected slack.attachment.fallback to contain the predefined message containg the pod and container name but it was not found.\nThe message was %v", podMsg)
			}

			if p.Reason != "" && !strings.Contains(msg, p.Reason) {
				t.Errorf("Expected slack.attachment.fallback to contain the supplied pod failure reason '%v' but it was not found", testCase.podReason)
			}

			if p.Message != "" && !strings.Contains(msg, p.Message) {
				t.Errorf("Expected slack.attachment.fallback to contain the supplied pod failure message '%v' but it was not found", testCase.podMessage)
			}

			if !strings.Contains(msg, p.StartedAt.Format(time.Stamp)) || !strings.Contains(msg, p.FinishedAt.Format(time.Stamp)) {
				t.Errorf("Expected slack.attachment.fallback to contain the correct time stamps '%v' and '%v' but one or more of these were not found", p.StartedAt.Format(time.Stamp), p.FinishedAt.Format(time.Stamp))
			}

			if msg != slackBody.Attachment[0].Field[0].Value {
				t.Errorf("Expected slack.attachment.fallback to match slack.attachment.field.value\n %v != %v", slackBody.Attachment[0].Fallback, slackBody.Attachment[0].Field[0].Value)
			}
		}

	}
}

// ExampleSTDOUTNotify is an Example that verifies that the notify function
// on STDOUT is printing the correct byte array to STDOUT
func ExampleSTDOUTNotify() {
	h := handler
	h = new(STDOUT)

	h.Notify([]byte("hello"))

	// Output:
	// hello
}
