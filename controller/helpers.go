package controller

import (
	"fmt"

	"gihutb.com/jxmoore/hubbub/models"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// getKubeClient pulls the InClusterConfig and returns the clientset
func getKubeClient() (*kubernetes.Clientset, error) {

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

// newNotification calls the methods on the interface that process a notification.
func newNotification(handler models.NotificationHandler, pod models.PodStatusInformation) error {

	msg, err := handler.BuildBody(pod)
	if err != nil {
		return fmt.Errorf("Error building body %v", err.Error())
	}

	err = handler.Notify(msg)
	if err != nil {
		return fmt.Errorf("Error sending notification %v", err.Error())
	}

	return nil
}
