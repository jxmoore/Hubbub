package init

import (
	"fmt"
	"log"

	"gihutb.com/jxmoore/hubbub/models"

	"gihutb.com/jxmoore/hubbub/watcher"
)

// BootStrap is the init function for the project, it is responsible for parsing the config,
// setting the handler, getting credentials and initiating the watch processes.
// Its exported as its called by Main.
func BootStrap(path string, envOnly bool) error {

	// load and parse inital config
	fmt.Println("Loading config...")
	config := &models.Config{}

	if !envOnly {
		if err := config.Load(path); err != nil {
			log.Fatal(err)
		}
	}

	config.LoadEnvVars()

	// I beleive that a NULL Namespace in Pods().Watch() will watch everything but because i havent tested it we just enforce it.
	if config.Namespace == "" {
		return fmt.Errorf("Please ensure the config has a Namespace specified")
	}

	fmt.Printf("Config loaded... \n %v\n", config)

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
		return fmt.Errorf("Error prepaing handler interface %v", err.Error())
	}

	// pull kubernetes incluster clientinfo
	client, err := watcher.GetKubeClient()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	watcher.StartWatcher(client, config, handler)

	return nil
}
