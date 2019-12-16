// Package models is the data layer, it contains all of structs used in the project and their associated methods.
package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
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
	STDOUT    bool   `json:"stdout,omitempty"`
	Debug     bool   `json:"debug,omitempty"`
	Self      string `json:"self,omitempty"`
	TimeCheck int    `json:"time"`
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

	if c.Namespace == "" && os.Getenv("HUBBUB_NAMESAPCE") != "" {
		c.Namespace = os.Getenv("HUBBUB_NAMESAPCE")
	}
	if c.Self == "" && os.Getenv("HUBBUB_SELF") != "" {
		c.Self = os.Getenv("HUBBUB_SELF")
	} else if c.Self == "" && os.Getenv("HUBBUB_SELF") == "" {
		c.Self = "Hubbub"
	}
	if !c.Debug && os.Getenv("HUBBUB_DEBUG") != "" {
		debug, err := strconv.ParseBool(os.Getenv("HUBBUB_DEBUG"))
		if err == nil {
			c.Debug = debug
		}
	}
	if !c.STDOUT && os.Getenv("HUBBUB_STDOUT") != "" {
		stdEnv, err := strconv.ParseBool(os.Getenv("HUBBUB_STDOUT"))
		if err == nil {
			c.STDOUT = stdEnv
		}
	}
	if c.TimeCheck == 0 && os.Getenv("HUBBUB_TIMECHECK") != "" {
		timeEnv, err := strconv.Atoi(os.Getenv("HUBBUB_TIMECHECK"))
		if err != nil {
			c.TimeCheck = 3
		} else {
			c.TimeCheck = timeEnv
		}
	} else if c.TimeCheck == 0 && os.Getenv("HUBBUB_TIMECHECK") == "" {
		c.TimeCheck = 3
	}
	if c.Slack.Channel == "" && os.Getenv("HUBBUB_CHANNEL") != "" {
		c.Slack.Channel = os.Getenv("HUBBUB_CHANNEL")
	}
	if c.Slack.WebHook == "" && os.Getenv("HUBBUB_WEBHOOK") != "" {
		c.Slack.WebHook = os.Getenv("HUBBUB_WEBHOOK")
	}
	if c.Slack.User == "" && os.Getenv("HUBBUB_USER") != "" {
		c.Slack.User = os.Getenv("HUBBUB_USER")
	} else if c.Slack.User == "" && os.Getenv("HUBBUB_USER") == "" {
		c.Slack.User = "Hubbub"
	}
	if c.Slack.Icon == "" && os.Getenv("HUBBUB_ICON") != "" {
		c.Slack.Icon = os.Getenv("HUBBUB_ICON")
	} else if c.Slack.Icon == "" && os.Getenv("HUBBUB_ICON") == "" {
		c.Slack.Icon = "https://www.sampalm.com/images/me.jpg"
	}
	if c.Slack.Title == "" && os.Getenv("HUBBUB_TITLE") != "" {
		c.Slack.Title = os.Getenv("HUBBUB_TITLE")
	} else if c.Slack.Title == "" && os.Getenv("HUBBUB_TITLE") == "" {
		c.Slack.Title = "There has been a pod error in production!"
	}

}
