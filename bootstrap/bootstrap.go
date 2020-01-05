package bootstrap

import (
	"fmt"

	"gihutb.com/jxmoore/hubbub/helpers"
	"gihutb.com/jxmoore/hubbub/models"

	"gihutb.com/jxmoore/hubbub/watcher"
)

// BootStrap is the init function for the project, it is responsible for parsing the config,
// setting the handler, getting credentials and initiating the watch processes.
// Its exported as its called by Main.
func BootStrap(path string, envOnly bool) error {

	// load and parse inital config
	fmt.Printf("Starting Hubbub...\n")
	config := &models.Config{}

	if !envOnly {
		if err := config.Load(path); err != nil {
			return fmt.Errorf("error loading config : \n%v", err)
		}
	}

	config.LoadEnvVars()

	// I beleive that a NULL Namespace in Pods().Watch() will watch everything but because i havent tested it we just enforce it.
	if config.Namespace == "" {
		return fmt.Errorf("please ensure the config has a Namespace specified")
	}

	helpers.DebugLog(config.Debug, "Configuration loaded...", config)

	// Setup the notifications interface
	var handler models.NotificationHandler
	if config.Notification.Handler == "slack" || config.Notification.Handler == "sl" {
		handler = new(models.Slack)
	} else if config.Notification.Handler == "appinsights" || config.Notification.Handler == "ai" || config.Notification.Handler == "applicationinsights" {
		handler = new(models.ApplicationInsights)
	} else {
		handler = new(models.STDOUT)
	}

	if err := handler.Init(config); err != nil {
		return fmt.Errorf("error prepaing handler interface : \n%v", err.Error())
	}

	// pull kubernetes incluster clientinfo
	client, err := helpers.GetKubeClient()
	if err != nil {
		return fmt.Errorf("error getting kubeclient info : \n%v", err.Error())
	}

	watcher.StartWatcher(client, config, handler)

	return nil
}
