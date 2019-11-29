package controller

import (
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"gihutb.com/jxmoore/hubbub/models"
	v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/kubernetes"
)

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

	fmt.Println("Listening on sigterm channel for termination...\n\n")

	// blocking while we wait for either signal on the channel, the podWatcher is running on his own go routine
	// so this just ensures everything runs until a signal is sent on the sigterm channel
	<-sigterm

	return nil
}

// podWatcher runs on a go routine that loops until we receive a signal on sigterm.
// It uses the resultchan in the watch interface to listen for events from Kubernetes and parses pod events, generating notifications for failed pods/containers.
func podWatcher(watcher watch.Interface, config *models.Config, handler models.NotificationHandler) {

	lastNotification := models.PodStatusInformation{}

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

			// ignore itself
			if config.Self != "" {
				if strings.Contains(strings.ToLower(pod.Name), strings.ToLower(config.Self)) {
					continue
				}
			}

			switch e.Type {

			// Modified is the only type we care about here
			// deletions and creation will be too noisey due to deployments
			case watch.Modified:

				if config.Debug {
					fmt.Printf("New pod change detected :\nPod : %v - Phase : %v\nMessage : %v - Reason : %v\nContainer info : \n%v\n",
						pod.Name, pod.Status.Phase, pod.Status.Message, pod.Status.Reason, pod.Status.ContainerStatuses)
				}

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
