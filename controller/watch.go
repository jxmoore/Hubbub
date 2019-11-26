package controller

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"gihutb.com/jxmoore/hubbub/models"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

// Not implemented... but what id like to do is compare images, pod etc... and verify we havent sent a notification for the same pod/container already.
var lastNotification = models.PodStatusInformation{}

// StartWatcher creates the watch.Interface{} that we listen on, kicks off the podWatcher go routine and listens for termination on a
// channel os.Signal. Its exported as its called
func startWatcher(kubeClient *kubernetes.Clientset, config *models.Config, handler models.NotificationHandler) error {

	// Create the watcher, nill listoptions should result in everything in NAMESPACE
	fmt.Println("Starting watcher...")
	watcher, err := kubeClient.CoreV1().Pods(config.Namespace).Watch(meta_v1.ListOptions{})
	if err != nil {
		return fmt.Errorf("Cannot create Pod event watcher, %v", err.Error())
	}

	fmt.Println("Listening on the watch channel for pod changes...")
	go podWatcher(watcher, config, handler)

	sigterm := make(chan os.Signal, 1)
	signal.Notify(sigterm, syscall.SIGTERM, syscall.SIGINT) // CTRL+C and term / exit

	fmt.Println("Listening on sigterm channel for termination...")

	// blocking while we wait for either signal on the channel, the podWatcher is running on his own go routine
	// so this just ensures everything runs until a signal is sent on the sigterm channel
	<-sigterm

	return nil
}

// podWatcher runs on a go routine that loops until we receive a signal on sigterm.
// It uses the resultchan in the watch interface to listen for events from Kubernetes and parses pod events, generating notifications for failed pods/containers.
func podWatcher(watcher watch.Interface, config *models.Config, handler models.NotificationHandler) {

	for {

		select {
		case e := <-watcher.ResultChan():

			if e.Object == nil {
				return
			}

			// Skip if not pod
			pod, ok := e.Object.(*v1.Pod)
			if !ok {
				continue
			}

			switch e.Type {

			// Modified is the only type we care about here
			// deletions and creation will be too noisey due to deployments
			case watch.Modified:

				if pod.DeletionTimestamp != nil {
					continue
				}

				podInformation := models.PodStatusInformation{}

				switch pod.Status.Phase {

				// Failing to start, encountered error etc....
				case v1.PodFailed:

					podInformation.Load(pod)
					podInformation.ConvertTime()

					if ok := podInformation.IsNew(lastNotification); ok {
						if err := newNotification(handler, podInformation); err != nil {
							fmt.Println(err.Error()) // non termintating
						} else {
							lastNotification = podInformation
						}
					}

				// Other issues
				default:

					podInformation.Load(pod)
					podInformation.ConvertTime()

					if ok := podInformation.IsNew(lastNotification); ok {
						if err := newNotification(handler, podInformation); err != nil {
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
