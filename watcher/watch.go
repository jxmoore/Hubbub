package watcher

import (
	"fmt"
	"strings"

	"gihutb.com/jxmoore/hubbub/helpers"
	"gihutb.com/jxmoore/hubbub/models"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// StartWatcher creates the watch.Interface{} that we listen on, kicks off the podWatcher go routine and listens for termination on a
// channel os.Signal. Its exported as its called
func StartWatcher(kubeClient *kubernetes.Clientset, config *models.Config, handler models.NotificationHandler) error {

	fmt.Printf("Starting the watcher...\n")
	for {

		// Create the watcher, nill listoptions should result in everything in NAMESPACE
		watcher, err := kubeClient.CoreV1().Pods(config.Namespace).Watch(meta_v1.ListOptions{})
		if err != nil {
			return fmt.Errorf("Cannot create Pod event watcher, %v", err.Error())
		}

		helpers.DebugLog(config.Debug, "Watcher created, starting podWatcher()")
		podWatcher(watcher, config, handler)

	}

	return nil
}

// podWatcher runs on a go routine that loops until we receive a signal on sigterm.
// It uses the resultchan in the watch interface to listen for events from Kubernetes and parses pod events, generating notifications for failed pods/containers.
func podWatcher(watcher watch.Interface, config *models.Config, handler models.NotificationHandler) {

	lastNotification := models.PodStatusInformation{}

	for {

		select {
		case e, open := <-watcher.ResultChan():

			if !open || e.Object == nil {
				helpers.DebugLog(config.Debug, "The channel has been closed, attempting to recreate the watcher")
				return
			}

			// Skip if not pod
			pod, ok := e.Object.(*v1.Pod)
			if !ok {
				continue
			}

			// ignore self
			if config.Self != "" {
				if strings.Contains(strings.ToLower(pod.Name), strings.ToLower(config.Self)) {
					helpers.DebugLog(config.Debug, "Detected and excluding a change, the pod is : "+pod.Name+". Skiping as it is matching the self attribute : '"+config.Self+"'.")
					continue
				}
			}

			switch e.Type {

			// Modified is the only type we care about here
			// deletions and creation will be too noisey due to deployments
			case watch.Modified:

				helpers.DebugLog(config.Debug, "New pod change detected : "+pod.Name+"\nMessage :"+pod.Status.Message+"\nReason : "+pod.Status.Reason, pod.Status.ContainerStatuses)

				if pod.DeletionTimestamp != nil {
					helpers.DebugLog(config.Debug, "Skipping pod : "+pod.Name+" as it was marked for deletion.")
					continue
				}

				podInformation := models.PodStatusInformation{}

				switch pod.Status.Phase {

				// Failing to start, encountered error etc....
				case v1.PodFailed:

					podInformation.Load(pod)
					podInformation.ConvertTime()

					if ok := podInformation.IsNew(lastNotification, config.TimeCheck); ok {
						if err := helpers.NewNotification(handler, podInformation); err != nil {
							fmt.Println(err.Error()) // non termintating
						} else {
							lastNotification = podInformation
						}
					}

				// Other issues
				default:

					podInformation.Load(pod)
					podInformation.ConvertTime()

					if ok := podInformation.IsNew(lastNotification, config.TimeCheck); ok {
						if err := helpers.NewNotification(handler, podInformation); err != nil {
							fmt.Println(err.Error()) // non termintating
						} else {
							lastNotification = podInformation
						}
					}
				}
			}
		}
	}
}
