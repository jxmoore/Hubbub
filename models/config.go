// Package models is the data layer, it contains all of structs used in the project and their associated methods.
package models

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
)

// Config is the struct that contains all of the hubbub config
type Config struct {
	Namespace    string `json:"namespace"`
	Labels       string `json:"labels"` // TODO, currently not implemented
	Debug        bool   `json:"debug"`
	Self         string `json:"self"`
	TimeCheck    int    `json:"time"`
	Notification struct {
		Handler string `json:"type"`
		// Slack specifics
		SlackWebHook string `json:"slackWebhook,omitempty"`
		SlackChannel string `json:"slackChannel,omitempty"`
		SlackTitle   string `json:"slackTitle,omitempty"`
		SlackUser    string `json:"slackUser,omitempty"`
		SlackIcon    string `json:"slackIcon,omitempty"`
		// Application Insights
		AppInsightsKey   string `json:"instrumentationKey,omitempty"`
		CustomEventTitle string `json:"customEventTitle.omitempty"`
	} `json:"notifications"`
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
		c.Notification.Handler = strings.ToLower(c.Notification.Handler)
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
	if c.Notification.SlackChannel == "" && os.Getenv("HUBBUB_CHANNEL") != "" {
		c.Notification.SlackChannel = os.Getenv("HUBBUB_CHANNEL")
	}
	if c.Notification.SlackWebHook == "" && os.Getenv("HUBBUB_WEBHOOK") != "" {
		c.Notification.SlackWebHook = os.Getenv("HUBBUB_WEBHOOK")
	}
	if c.Notification.SlackUser == "" && os.Getenv("HUBBUB_USER") != "" {
		c.Notification.SlackUser = os.Getenv("HUBBUB_USER")
	} else if c.Notification.SlackUser == "" && os.Getenv("HUBBUB_USER") == "" {
		c.Notification.SlackUser = "Hubbub"
	}
	if c.Notification.SlackIcon == "" && os.Getenv("HUBBUB_ICON") != "" {
		c.Notification.SlackIcon = os.Getenv("HUBBUB_ICON")
	} else if c.Notification.SlackIcon == "" && os.Getenv("HUBBUB_ICON") == "" {
		c.Notification.SlackIcon = "https://www.sampalm.com/images/me.jpg"
	}
	if c.Notification.SlackTitle == "" && os.Getenv("HUBBUB_TITLE") != "" {
		c.Notification.SlackTitle = os.Getenv("HUBBUB_TITLE")
	} else if c.Notification.SlackTitle == "" && os.Getenv("HUBBUB_TITLE") == "" {
		c.Notification.SlackTitle = "There has been a pod error in production!"
	}

}
