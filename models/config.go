// Package models is the data layer, it contains all of structs used in the project and their associated methods.
package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
)

// Config is the struct that contains all of the hubbub config
type Config struct {
	Namespace string `json:"namespace,omitempty"`
	Labels    string `json:"labels,omitempty"` // TODO, currently not implemented
	Slack     struct {
		WebHook string `json:"webhook"`
		Channel string `json:"channel"`
		Title   string `json:"title,omitempty"`
		User    string `json:"user,omitempty"`
		Icon    string `json:"icon,omitempty"`
	} `json:"slack"`
	STDOUT bool `json:"stdout,omitempty"`
	Debug  bool `json:"debug,omitempty"`
}

// Load attempts to read the config file and unmarshel it into 'c'
func (c *Config) Load(configFile string) error {

	file, err := os.Open(configFile)
	if err != nil {
		return fmt.Errorf("Unable to open config file %v. %v ", configFile, err.Error())
	}

	defer file.Close()

	content, err := ioutil.ReadAll(file)
	if err != nil {
		return fmt.Errorf("Error reading file %v. %v ", configFile, err.Error())
	}

	if len(content) != 0 {
		return json.Unmarshal(content, c)
	}

	return nil
}

// LoadEnvVars will pull the associated enviroment variables and assign them to 'c' if they
// are missing from 'c' and present as env variables. Some values, such as the icon, have a default value which is
// assigned here if 'c' and the env variable is nil.
func (c *Config) LoadEnvVars() {

	if c.Namespace == "" && os.Getenv("NAMESAPCE") != "" {
		c.Namespace = os.Getenv("NAMESAPCE")
	}
	if c.Slack.Channel == "" && os.Getenv("SLACK_CHANNEL") != "" {
		c.Slack.Channel = os.Getenv("SLACK_CHANNEL")
	}
	if c.Slack.WebHook == "" && os.Getenv("SLACK_WEBHOOK") != "" {
		c.Slack.WebHook = os.Getenv("SLACK_WEBHOOK")
	}
	if c.Slack.User == "" && os.Getenv("SLACK_USER") != "" {
		c.Slack.User = os.Getenv("SLACK_USER")
	} else if c.Slack.User == "" && os.Getenv("SLACK_USER") == "" {
		c.Slack.User = "Hubbub"
	}
	if c.Slack.Icon == "" && os.Getenv("SLACK_ICON") != "" {
		c.Slack.Icon = os.Getenv("SLACK_ICON")
	} else if c.Slack.Icon == "" && os.Getenv("SLACK_ICON") == "" {
		c.Slack.Icon = "https://www.sampalm.com/images/me.jpg"
	}
	if c.Slack.Title == "" && os.Getenv("SLACK_TITLE") != "" {
		c.Slack.Title = os.Getenv("SLACK_TITLE")
	} else if c.Slack.Title == "" && os.Getenv("SLACK_TITLE") == "" {
		c.Slack.Title = "There has been a pod error in production!"
	}

}
