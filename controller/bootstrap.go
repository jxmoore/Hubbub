package controller

import (
	"fmt"
	"log"

	"gihutb.com/jxmoore/hubbub/models"
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

	// pull kubernetes incluster clientinfo
	client, err := getKubeClient()
	if err != nil {
		return fmt.Errorf(err.Error())
	}

	// Setup the notifications interface
	var handler models.NotificationHandler
	if config.Slack.WebHook == "" && config.STDOUT {
		fmt.Println("Using STDOUT...")
		handler = new(models.STDOUT)
	} else {
		fmt.Println("Using Slack...")
		handler = new(models.Slack)
	}

	if err := handler.Init(config); err != nil {
		return fmt.Errorf("Error prepaing handler interface %v", err.Error())
	}

	startWatcher(client, config, handler)

	return nil
}
