package controller

import (
	"gihutb.com/jxmoore/hubbub/models"
)

// pod is a package wide PodStatusInformation{} used in all of the controller tests as a base.
var pod = models.PodStatusInformation{
	Namespace:     "hubbub",
	PodName:       "hubbub",
	ContainerName: "hubbub",
	Image:         "hubbub",
	ExitCode:      2,
	Reason:        "hubbub",
	Message:       "hubbub",
}

func ExampleNewNotification() {

	var handler models.NotificationHandler
	handler = new(models.STDOUT)
	p := pod
	p.Namespace = "hubbub-Testin"

	newNotification(handler, p)

	// Output:
	// {"Namespace":"hubbub-Testin","PodName":"hubbub","ContainerName":"hubbub","Image":"hubbub","StartedAt":"0001-01-01T00:00:00Z","FinishedAt":"0001-01-01T00:00:00Z","ExitCode":2,"Reason":"hubbub","Message":"hubbub","Seen":"0001-01-01T00:00:00Z"}

}
