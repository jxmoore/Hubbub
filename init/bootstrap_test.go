package init

import (
	"encoding/json"
	"os"
	"testing"

	"gihutb.com/jxmoore/hubbub/models"
)

// configFile is a package wide Config{}
var configFile = models.Config{
	Namespace: "jomo",
	STDOUT:    false,
}

func TestBootStrap(t *testing.T) {

	// Currently no way of testing locally with kubeconfig so this will fail with the following :
	knownIssue := "Can not get kubernetes config: unable to load in-cluster configuration, KUBERNETES_SERVICE_HOST and KUBERNETES_SERVICE_PORT must be defined"

	// Our env variables that will be passed in
	envVariables := map[string]string{
		"NAMESAPCE":     "hubbubTest",
		"SLACK_CHANNEL": "#techGeneral",
		"SLACK_WEBHOOK": "google.com",
		"SLACK_USER":    "Hubub",
		"SLACK_ICON":    "google.com/icon/happyfeet.png",
		"SLACK_TITLE":   "A pod is suffering many woes!",
	}

	testSuite := map[string]struct {
		// Values for Conf{}
		namespace string
		webhook   string
		channel   string
		STDOUT    bool
		Debug     bool
		Self      string
		// Where to store Conf{}
		filePath string
		// Use only env variables (passed to BootStrap()) and the map of the values
		useEnv bool
		envVar map[string]string
		// The expected error response as a string (err.Error())
		errorResponse string
		// Clean up the configs after the test
		clean bool
	}{
		"Create a config with Default namespace and pass it into BootStrap": {
			namespace:     "Default",
			webhook:       "https://github.com/jxmoore/Hubbub/tree/develop/models",
			channel:       "#Tech_General",
			STDOUT:        false,
			Debug:         true,
			filePath:      "./testConf1.json",
			Self:          "Hubbub",
			clean:         true,
			useEnv:        false,
			errorResponse: knownIssue,
		},
		"Create a config with Secret namespace and pass it into BootStrap": {
			namespace:     "Secret",
			webhook:       "https://duckduckgo.com",
			channel:       "#Tech_InfoSec",
			STDOUT:        false,
			Self:          "Hubbub",
			Debug:         true,
			filePath:      "./testConf2.json",
			clean:         true,
			useEnv:        false,
			errorResponse: knownIssue,
		},
		"Create a config with testo namespace and pass it into BootStrap": {
			namespace:     "testo",
			webhook:       "https://github.com",
			channel:       "#random",
			Self:          "Hubbub",
			STDOUT:        true,
			Debug:         false,
			filePath:      "./testConf3.json",
			clean:         true,
			useEnv:        false,
			errorResponse: knownIssue,
		},
		"Use Env Variables": {
			filePath:      "./noRealFile/Here.log",
			useEnv:        true,
			envVar:        envVariables,
			errorResponse: knownIssue,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		if testCase.useEnv {

			for k, v := range testCase.envVar {
				os.Setenv(k, v)
			}

		} else {
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
					t.Fatalf("Temp conf file already exists!")
				} else {
					t.Fatalf("Error creating file %v", err.Error())
				}
			}

			_, err = file.Write(content)
			if err != nil {
				t.Fatalf("Error writing to %v %v", testCase.filePath, err.Error())
			}
			// no defer as it does not close the file quick enough if testCase.clean == true
			file.Close()

		}

		if err := BootStrap(testCase.filePath, testCase.useEnv); err != nil {
			if err.Error() != testCase.errorResponse {
				t.Errorf("Expected BootStrap to return the error %v\nReceived %v", testCase.errorResponse, err.Error())
			}
		}

	}

	for _, testCase := range testSuite {
		if testCase.clean {
			if err := os.Remove(testCase.filePath); err != nil {
				t.Logf("Error on deletion  %v", err) // tests should not fail on cleanup
			}

		}
	}

}
