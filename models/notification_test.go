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
		// config.notifications.type :
		notificationType string
		// slack specific
		webhook string
		channel string
		// app insights specific
		eventTitle string
		key        string
		// the expected response (error)
		expectedResponse string
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
		"(s *ApplicationInsights) Init() will throw an error due to missing fields": {
			notificationType: "ai",
			expectedResponse: "missing instrumentation key",
		},
		"(s *ApplicationInsights) Init() will return nil": {
			notificationType: "ai",
			eventTitle:       "#broTalk",
			key:              "as8932OsdAS89DQR54FDas8932OsdAS89DQR54FD",
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		notification := handler

		if testCase.notificationType == "slack" {
			notification = new(Slack)
		} else if testCase.notificationType == "ai" {
			notification = new(ApplicationInsights)
		} else {
			notification = new(STDOUT)
		}

		fakeConf := configFile // from config_test.go
		fakeConf.Notification.SlackWebHook = testCase.webhook
		fakeConf.Notification.SlackChannel = testCase.channel
		fakeConf.Notification.AppInsightsKey = testCase.key
		fakeConf.Notification.CustomEventTitle = testCase.eventTitle

		if err := notification.Init(&fakeConf); err != nil {
			if testCase.expectedResponse != err.Error() {
				t.Fatalf("Expected an error code of %v and received %v", testCase.expectedResponse, err)
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
	configFile.Notification.SlackUser = "hubbub"
	configFile.Notification.SlackIcon = "https://www.sampalm.com/images/me.jpg"
	configFile.Notification.SlackChannel = "Testing"
	configFile.Notification.SlackWebHook = "https://www.sampalm.com/images/me.jpg"

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
			podReason:        "Stuffs",
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

		p := TestPod // from kube_test.go
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
		c.Notification.SlackWebHook = "google.com"
		c.Notification.SlackTitle = "Oh no!"

		bodyHandler.Init(&c)
		p.ConvertTime()
		msgInBytes, _ := BuildBody(bodyHandler, p)

		if testCase.notificationType == "stdout" {

			pCheck := PodStatusInformation{}
			json.Unmarshal(msgInBytes.body, &pCheck)

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
			json.Unmarshal(msgInBytes.body, &slackBody)
			msg := slackBody.Attachment[0].Fallback

			if slackBody.Attachment[0].Color != "danger" {
				t.Errorf("Expected Notification.Slackattachment.color == danger but received %v", slackBody.Attachment[0].Color)
			}

			if slackBody.Attachment[0].Title != c.Notification.SlackTitle {
				t.Errorf("Expected Notification.Slackattachment.title to match the title supplied in the handler init but received %v", slackBody.Attachment[0].Title)
			}

			if !strings.Contains(msg, errorDetails) {
				t.Errorf("Expected Notification.Slackattachment.fallback to contain the error details but the string %v was not found.", errorDetails)
			}

			if !strings.Contains(msg, podMsg) {
				t.Errorf("Expected Notification.Slackattachment.fallback to contain the predefined message containg the pod and container name but it was not found.\nThe message was %v", podMsg)
			}

			if p.Reason != "" && !strings.Contains(msg, p.Reason) {
				t.Errorf("Expected Notification.Slackattachment.fallback to contain the supplied pod failure reason '%v' but it was not found. %v", testCase.podReason, msg)
			}

			if p.Message != "" && !strings.Contains(msg, p.Message) {
				t.Errorf("Expected Notification.Slackattachment.fallback to contain the supplied pod failure message '%v' but it was not found", testCase.podMessage)
			}

			if !strings.Contains(msg, p.StartedAt.Format(time.Stamp)) || !strings.Contains(msg, p.FinishedAt.Format(time.Stamp)) {
				t.Errorf("Expected Notification.Slackattachment.fallback to contain the correct time stamps '%v' and '%v' but one or more of these were not found", p.StartedAt.Format(time.Stamp), p.FinishedAt.Format(time.Stamp))
			}

			if msg != slackBody.Attachment[0].Field[0].Value {
				t.Errorf("Expected Notification.Slackattachment.fallback to match Notification.Slackattachment.field.value\n %v != %v", slackBody.Attachment[0].Fallback, slackBody.Attachment[0].Field[0].Value)
			}
		}

	}
}

// ExampleSTDOUTNotify is an Example that verifies that the notify function
// on STDOUT is printing the correct byte array to STDOUT
func ExampleSTDOUTNotify() {

	h := handler
	h = new(STDOUT)
	nDetails := NotificationDetails{body: []byte("hello")}
	h.Notify(nDetails)

	// Output:
	// hello
}
