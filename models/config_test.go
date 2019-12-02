package models

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

// configFile is a package wide Config{} used in all of the model tests as a base.
var configFile = Config{
	Namespace: "jomo",
	STDOUT:    false,
}

// TestLoadConfig tests the Load() method on Config.
// It does this by creating a config file using a Config{} and verifying that the struct used
// to write the json config file matches the struct returned from the Config.Load() method
func TestConfigLoad(t *testing.T) {

	testSuite := map[string]struct {
		namespace string
		webhook   string
		channel   string
		STDOUT    bool
		Debug     bool
		Self      string
		filePath  string
		// Delete the config files or let them persist on disk
		clean bool
	}{
		"Created config with Default namespace should be identical to loaded config": {
			namespace: "Default",
			webhook:   "https://github.com/jxmoore/Hubbub/tree/develop/models",
			channel:   "#Tech_General",
			STDOUT:    false,
			Debug:     true,
			filePath:  "./testConf1.json",
			Self:      "Default",
			clean:     true,
		},
		"Created config with Secret namespace should be identical to loaded config": {
			namespace: "Secret",
			webhook:   "https://duckduckgo.com",
			channel:   "#Tech_InfoSec",
			STDOUT:    false,
			Self:      "infosec",
			Debug:     true,
			filePath:  "./testConf2.json",
			clean:     true,
		},
		"Created config with testo namespace should be identical to loaded config": {
			namespace: "testo",
			webhook:   "https://bitbucket.com",
			channel:   "#random",
			Self:      "stuff",
			STDOUT:    true,
			Debug:     false,
			filePath:  "./testConf3.json",
			clean:     true,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		configFile.Namespace = testCase.namespace
		configFile.Slack.WebHook = testCase.webhook
		configFile.Slack.Channel = testCase.channel
		configFile.STDOUT = testCase.STDOUT
		configFile.Debug = testCase.Debug

		content, err := json.Marshal(configFile)
		if err != nil {
			t.Errorf("Error marshalling JSON %v", err.Error())
		}

		file, err := os.OpenFile(testCase.filePath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0666)
		if err != nil {
			if os.IsExist(err) {
				t.Errorf("Temp conf file already exists!")
			} else {
				t.Errorf("Error creating file %v", err.Error())
			}
		}

		_, err = file.Write(content)
		if err != nil {
			t.Errorf("Error writing to %v %v", testCase.filePath, err.Error())
		}

		// no defer as it does not close the file quick enough if testCase.clean == true
		file.Close()

		newConf := Config{}
		if ok := newConf.Load(testCase.filePath); ok == nil {

			if reflect.DeepEqual(newConf, configFile) {
				t.Logf("Config structs match")
			} else {
				t.Errorf("Struct returned from Load() does not match the test struct!")
			}

		} else {
			t.Errorf("Error on Load() %v", ok)
		}

	}

	for _, testCase := range testSuite {
		if testCase.clean {
			if err := os.Remove(testCase.filePath); err != nil {
				t.Errorf("Error on deletion  %v", err)
			}

		}
	}

}

// TestVarLoad() Creates the expected env variables found in LoadEnvVars() and then calls the method on an empty Config{}
// assuming everything is working as expected the Config should contain the values for the env variables or the default values
// based on the test case.
func TestVarLoad(t *testing.T) {

	testSuite := map[string]struct {
		slackChannel string
		webhook      string
		user         string
		icon         string
		title        string
		namespace    string
	}{
		"Config should be populated with values derived from ENV variables #1": {
			slackChannel: "hubbub",
			webhook:      "hubbub",
			user:         "hubbub",
			icon:         "hubbub",
			title:        "hubbub",
			namespace:    "hubbub",
		},
		"Config should be populated with values derived from ENV variables #2": {
			slackChannel: "alerts",
			webhook:      "alerts",
			user:         "alerts",
			icon:         "alerts",
			title:        "alerts",
			namespace:    "alerts",
		},
		"Config should be populated with values derived from the defaults in LoadEnvVars #3": {
			slackChannel: "alerts",
			webhook:      "alerts",
			user:         "",
			icon:         "",
			title:        "",
			namespace:    "alerts",
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		c := Config{}
		os.Setenv("NAMESAPCE", testCase.namespace)
		os.Setenv("SLACK_CHANNEL", testCase.slackChannel)
		os.Setenv("SLACK_WEBHOOK", testCase.webhook)
		os.Setenv("SLACK_USER", testCase.user)
		os.Setenv("SLACK_ICON", testCase.icon)
		os.Setenv("SLACK_TITLE", testCase.title)
		c.LoadEnvVars()

		if testCase.user == "" {
			testCase.user = "Hubbub"
		}
		if testCase.icon == "" {
			testCase.icon = "https://www.sampalm.com/images/me.jpg"
		}
		if testCase.title == "" {
			testCase.title = "There has been a pod error in production!"
		}

		if c.Slack.Channel != testCase.slackChannel {
			t.Errorf("Expected the slack channel in the config to match the testcase %v but received %v", testCase.slackChannel, c.Slack.Channel)
		}
		if c.Slack.WebHook != testCase.webhook {
			t.Errorf("Expected the slack webhook in the config to match the testcase %v but received %v", testCase.webhook, c.Slack.WebHook)
		}
		if c.Namespace != testCase.namespace {
			t.Errorf("Expected the namespace in the config to match the testcase %v but received %v", testCase.namespace, c.Namespace)
		}
		if testCase.user != c.Slack.User {
			t.Errorf("Expected the slack user in the config to match the testcase %v but received %v", testCase.user, c.Slack.User)
		}
		if testCase.icon != c.Slack.Icon {
			t.Errorf("Expected the slack icon in the config to match the testcase %v but received %v", testCase.icon, c.Slack.Icon)
		}
		if testCase.title != c.Slack.Title {
			t.Errorf("Expected the slack title message in the config to match the testcase %v but received %v", testCase.title, c.Slack.Title)
		}
	}

}
