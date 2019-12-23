package watcher

import (
	"fmt"

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

	msg := models.BuildBody(handler, pod)
	if err := handler.Notify(msg); err != nil {
		return fmt.Errorf("Error sending notification %v", err.Error())
	}

	return nil
}
