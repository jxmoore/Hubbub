package models

import (
	"encoding/json"
	"os"
	"reflect"
	"testing"
)

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
		filePath  string
		// Delete the config files or let them persist on disk
		clean bool
	}{
		"Created config with Default namespace should be identical to loaded config": {
			namespace: "Default",
			webhook:   "https://github.com/jxmoore/Hubbub/tree/develop/models",
			channel:   "#Tech_General",
			STDOUT:    false,
			filePath:  "./testConf1.json",
			clean:     true,
		},
		"Created config with Secret namespace should be identical to loaded config": {
			namespace: "Secret",
			webhook:   "https://duckduckgo.com",
			channel:   "#Tech_InfoSec",
			STDOUT:    false,
			filePath:  "./testConf2.json",
			clean:     true,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		configFile.Namespace = testCase.namespace
		configFile.Slack.WebHook = testCase.webhook
		configFile.Slack.Channel = testCase.channel
		configFile.STDOUT = testCase.STDOUT

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
