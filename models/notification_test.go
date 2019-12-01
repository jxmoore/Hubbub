package models

import (
	"testing"
)

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

		var handler NotificationHandler
		handler = new(STDOUT)

		if testCase.notificationType == "slack" {
			handler = new(Slack)
		}

		fakeConf := configFile // from config_test.go
		fakeConf.Slack.WebHook = testCase.webhook
		fakeConf.Slack.Channel = testCase.channel

		if err := handler.Init(&fakeConf); err != nil {
			if testCase.expectedResponse != err.Error() {
				t.Errorf("Expected an error code of %v and received %v", testCase.expectedResponse, err)
			} else {
				t.Logf("Received the correct error response from Init()")
			}
		}
	}
}
