package helpers

import (
	"fmt"
	"time"

	"gihutb.com/jxmoore/hubbub/models"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// GetKubeClient pulls the InClusterConfig and returns the clientset
func GetKubeClient() (*kubernetes.Clientset, error) {

	// Get the config
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("Can not get kubernetes config: %v", err)
	}

	// Get the client info
	kubeClient, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("Can not get kubernetes config: %v", err)
	}

	fmt.Println("Kube credentials pulled...")
	return kubeClient, nil
}

// NewNotification calls the methods on the interface that process a notification.
func NewNotification(handler models.NotificationHandler, pod models.PodStatusInformation) error {

	msg, err := models.BuildBody(handler, pod)
	if err != nil {
		return err
	}

	if err := handler.Notify(msg); err != nil {
		return fmt.Errorf("Error sending notification %v", err.Error())
	}

	return nil
}

// DebugLog is a helper function that prints one or more items to the console if the debug flag is flipped.
// It takes the empty interface as structs from other packages (namely the models pacakage) may be passed in; however, 
// generally speaking only strings are expected.
func DebugLog(debug bool, input ...interface{}) {

	if !debug {
		return
	}

	for _, i := range input {
		fmt.Printf("DEBUG [%v]: %T : %v\n", time.Now().Format("January 2, 2006"), i, i)
	}

}
