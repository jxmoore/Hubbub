package models

import (
	"encoding/json"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"testing"
	"time"
)

// configFile is a package wide Config{} used in all of the model tests as a base.
var configFile = Config{
	Namespace: "jomo",
	TimeZone:  "America/New_York",
}

// TestLoadConfig tests the Load() method on Config.
// It does this by creating a config file using a Config{} and verifying that the struct used
// to write the json config file matches the struct returned from the Config.Load() method
func TestConfigLoad(t *testing.T) {

	testSuite := map[string]struct {
		// Values for the config
		namespace string
		webhook   string
		channel   string
		Debug     bool
		Self      string
		filePath  string
		timeCheck int
		// Delete the config files or let them persist on disk
		clean bool
	}{
		"Created config with Default namespace should be identical to loaded config": {
			namespace: "Default",
			webhook:   "https://github.com/jxmoore/Hubbub/tree/develop/models",
			channel:   "#Tech_General",
			Debug:     true,
			filePath:  "./testConf1.json",
			Self:      "Default",
			timeCheck: 6,
			clean:     true,
		},
		"Created config with Secret namespace should be identical to loaded config": {
			namespace: "Secret",
			webhook:   "https://duckduckgo.com",
			channel:   "#Tech_InfoSec",
			Self:      "infosec",
			Debug:     true,
			filePath:  "./testConf2.json",
			timeCheck: 18,
			clean:     true,
		},
		"Created config with testo namespace should be identical to loaded config": {
			namespace: "testo",
			webhook:   "https://bitbucket.com",
			channel:   "#random",
			Self:      "stuff",
			timeCheck: 3,
			Debug:     false,
			filePath:  "./testConf3.json",
			clean:     true,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		configFile.Namespace = testCase.namespace
		configFile.Notification.SlackWebHook = testCase.webhook
		configFile.Notification.SlackChannel = testCase.channel
		configFile.Self = testCase.Self
		configFile.Debug = testCase.Debug
		configFile.TimeCheck = testCase.timeCheck
		configFile.TimeLocation, _ = time.LoadLocation(configFile.TimeZone)

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

		newConf := Config{}
		if ok := newConf.Load(testCase.filePath); ok == nil {

			if reflect.DeepEqual(newConf, configFile) {
				t.Logf("Config structs match")
			} else {
				t.Errorf("Struct returned from Load() does not match the test struct!, %v != %v", newConf, configFile)
			}

		} else {
			t.Errorf("Error on Load() %v", ok)
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

// TestVarLoad() Sets a series of env variables based on the keys/values found in a map and then calls the LoadEnvVar() method on an empty Config{}.
// Assuming everything is working as expected the Config should contain the values from the map or the default values based on the test case.
//
// This test is using reflect to loop over the fields in the struct and using their name/values to create enviroment variables.
// While slightly more complex than the past test version, this version of the test is much more exstensible.
func TestVarLoad(t *testing.T) {

	var envVariables = map[string]string{
		"NAMESAPCE": "hubbubTest",
		"CHANNEL":   "#techGeneral",
		"WEBHOOK":   "google.com",
	}

	// The fields match the suffix for a respective ENV variable. For example Namespace = HUBBUB_NAMESPACE
	// during the test when os.SetEnv() is called the key is updated to include the prefix
	testSuite := map[string]struct {
		NAMESAPCE string
		CHANNEL   string
		WEBHOOK   string
		USER      string
		ICON      string
		TITLE     string
		SELF      string
		TIMECHECK string
		DEBUG     string
		STDOUT    string
	}{
		"Config should be populated with values derived from ENV variables #1": {
			CHANNEL:   "#kubes",
			NAMESAPCE: "hubbub",
			WEBHOOK:   "slack.com/webhooks",
			USER:      "jomo",
			ICON:      "face.png",
			TITLE:     "Somethings awry",
			SELF:      "man",
			TIMECHECK: "8",
			DEBUG:     "true",
			STDOUT:    "true",
		},
		"Config should be populated with values derived from ENV variables #2": {
			CHANNEL:   "#bub-Tub",
			NAMESAPCE: "bubub",
			WEBHOOK:   "slack.com/stuff/hook",
			USER:      "jomo",
			ICON:      "hubbubTest.png",
			TITLE:     "Prod problems",
			SELF:      "jakey",
			TIMECHECK: "15",
			DEBUG:     "false",
			STDOUT:    "false",
		},
		"Config should be populated with values derived from the defaults in LoadEnvVars #3": {
			CHANNEL:   "#hubbub",
			NAMESAPCE: "default",
			WEBHOOK:   "badURL",
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		c := Config{}

		envVars := envVariables
		valOf := reflect.ValueOf(testCase)
		typeOf := valOf.Type()

		for i := 0; i < valOf.NumField(); i++ {
			element := valOf.Field(i)
			key := typeOf.Field(i).Name
			if element.Type().String() == "string" {
				envVars[key] = fmt.Sprintf("%v", element.Interface())
			}
		}

		for k, v := range envVars {
			// each env variable is prefixed with HUBBUB_
			os.Setenv("HUBBUB_"+k, v)
		}

		c.LoadEnvVars()

		// the defaults from LoadEnvVars()
		if envVariables["USER"] == "" {
			envVariables["USER"] = "Hubbub"
		}
		if envVariables["SELF"] == "" {
			envVariables["SELF"] = "Hubbub"
		}
		if envVariables["ICON"] == "" {
			envVariables["ICON"] = "https://www.sampalm.com/images/me.jpg"
		}
		if envVariables["TITLE"] == "" {
			envVariables["TITLE"] = "There has been a pod error in production!"
		}
		if envVariables["TIMECHECK"] == "" {
			envVariables["TIMECHECK"] = "3"
		}
		if envVariables["DEBUG"] == "" {
			envVariables["DEBUG"] = "false"
		}
		if envVariables["STDOUT"] == "" {
			envVariables["STDOUT"] = "false"
		}
		debug, _ := strconv.ParseBool(envVariables["DEBUG"])
		timeEnv, _ := strconv.Atoi(envVariables["TIMECHECK"])

		// c should be populated from the env variables on LoadEnvVars(), so we cross check the fields in 'C' against our map values
		if c.Notification.SlackChannel != envVariables["CHANNEL"] {
			t.Errorf("Expected the slack channel in the config to match the testcase value '%v' but received '%v'", envVariables["CHANNEL"], c.Notification.SlackChannel)
		}
		if c.Notification.SlackWebHook != envVariables["WEBHOOK"] {
			t.Errorf("Expected the slack webhook in the config to match the testcase value '%v' but received '%v'", envVariables["WEBHOOK"], c.Notification.SlackWebHook)
		}
		if c.Namespace != envVariables["NAMESAPCE"] {
			t.Errorf("Expected the namespace in the config to match the testcase value '%v' but received '%v'", envVariables["NAMESAPCE"], c.Namespace)
		}
		if c.Notification.SlackUser != envVariables["USER"] {
			t.Errorf("Expected the slack user in the config to match the testcase value '%v' but received '%v'", envVariables["USER"], c.Notification.SlackUser)
		}
		if c.Notification.SlackIcon != envVariables["ICON"] {
			t.Errorf("Expected the slack icon in the config to match the testcase value '%v' but received '%v'", envVariables["ICON"], c.Notification.SlackIcon)
		}
		if c.Notification.SlackTitle != envVariables["TITLE"] {
			t.Errorf("Expected the slack title message in the config to match the testcase value '%v' but received '%v'", envVariables["TITLE"], c.Notification.SlackTitle)
		}
		if c.Self != envVariables["SELF"] {
			t.Errorf("Expected the SELF value in the config to match the testcase value '%v' but received '%v'", envVariables["SELF"], c.Self)
		}
		if c.TimeCheck != timeEnv {
			t.Errorf("Expected the TimeCheck value in the config to match the testcase value '%v' but received '%v'", timeEnv, c.TimeCheck)
		}
		if c.Debug != debug {
			t.Errorf("Expected the Debug value in the config to match the testcase value '%v' but received '%v'", debug, c.Debug)
		}
	}
}
