package models

import (
	"reflect"
	"time"

	v1 "k8s.io/api/core/v1"
)

// PodStatusInformation is a small struct that stores info about pod and container changes.
type PodStatusInformation struct {
	Namespace     string
	PodName       string
	ContainerName string
	Image         string
	StartedAt     time.Time
	FinishedAt    time.Time
	ExitCode      int
	Reason        string
	Message       string
	Seen          time.Time
}

// Load takes a pod info and then loads the attributes of them into the PodStatusInformation struct
func (p *PodStatusInformation) Load(pod *v1.Pod) {

	p.Namespace = pod.Namespace
	p.StartedAt = pod.CreationTimestamp.Time
	p.PodName = pod.Name
	p.Message = pod.Status.Message
	p.Seen = time.Now()

	if len(pod.Status.ContainerStatuses) == 0 && pod.Status.Phase == v1.PodFailed { // skip pending in default case

		p.FinishedAt = pod.CreationTimestamp.Time
		p.ExitCode = -1
		p.Reason = pod.Status.Reason
		p.Image = "Unknown"
		p.ContainerName = "Unknown"

		if len(pod.Spec.Containers) > 0 {
			p.Image = pod.Spec.Containers[0].Image
			p.ContainerName = pod.Spec.Containers[0].Name
		}

	} else {

		for _, cst := range pod.Status.ContainerStatuses {

			// Skipping the terminated container
			if cst.State.Terminated == nil {
				continue
			}

			// If we land in default we need to ensure that we dont alert on good pods
			if cst.State.Terminated.Reason != "Completed" {
				p.FinishedAt = cst.State.Terminated.FinishedAt.Time
				p.Image = cst.Image
				p.ContainerName = cst.Name
				p.ExitCode = int(cst.State.Terminated.ExitCode)
				p.Reason = cst.State.Terminated.Reason
				p.Message = cst.State.Terminated.Message
				break
			}
		}
	}
}

// IsNew compares fields in p with the ones passed in on lastSeen. The purpose is to validate
// that a instance of the struct is new, if it is a repeat or if its close enough to be a repeat.
// This is done to cut down on redundant messages generated in Watch(). For e.g. a pod that terminates
// shortly after startup due to an error in program.cs and restarts will generate constant alerts as
// it constantly goes up and down but we only want the first (or one thats older than 'X').
func (p PodStatusInformation) IsNew(lastSeen PodStatusInformation, timeSince int) bool {

	// assume not a failure
	if p.Image == "" || p.ContainerName == "" || p.FinishedAt.IsZero() {
		return false
	}

	p.ConvertTime()
	lastSeen.ConvertTime()

	// Check to see if its been over 'x' minutes, if so its new yet again.
	if ok := p.timeCheck(lastSeen, timeSince); ok {
		return true
	}

	// Identical
	if reflect.DeepEqual(p, lastSeen) {

		return false
	}

	if p.PodName == lastSeen.PodName && p.ContainerName == lastSeen.PodName {

		return false
	}

	// Same pod, same start time
	if p.PodName == lastSeen.PodName && p.StartedAt == lastSeen.StartedAt {

		return false
	}

	// same container, same exit code
	if p.ContainerName == lastSeen.ContainerName && p.ExitCode == lastSeen.ExitCode {

		return false
	}

	// same container, same exit code
	if p.PodName == lastSeen.PodName && p.ExitCode == lastSeen.ExitCode {

		return false
	}

	return true
}

// ConvertTime converts all of the times found in p to local (EST). This is in place because some
// users host their Kubeernetes clusters in cloud enviroments where the local timezone does not match
// the end users.
//
// TODO : The location should be exsposed in the config allowing other time zones.
func (p *PodStatusInformation) ConvertTime() {

	zone, err := time.LoadLocation("America/New_York")
	if err == nil {
		p.FinishedAt = p.FinishedAt.In(zone)
		p.StartedAt = p.StartedAt.In(zone)
	} // on err we just maintin the original times

}

// ExitCodeLookup tries to associate the int exit code in 'p' to a string that describes the exit code.
// E.G. 139 = Segmentation fault
func (p PodStatusInformation) ExitCodeLookup() string {

	exitCodes := map[int]string{
		139: "Segmentation fault.",
		143: "The container received s SIGTERM.",
		137: "The container received a SIGKILL.",
		127: "Command not found.",
		130: "Container terminated.",
		126: "There was a error regardging permissions or the container could not be invoked.",
		125: "The Docker Run command has failed.",
		1:   "Application Error.",
	}

	if i, ok := exitCodes[p.ExitCode]; ok {
		return i
	}

	return ""

}

// timeCheck checks to see if a pod was seen more than 'x' minutes ago. This is acheived
// by diffing the 'seen' values in both struct ('p' and 'lastSeen') and then seeing if it
// exceeds the timeSince value provided in the config.
func (p PodStatusInformation) timeCheck(lastSeen PodStatusInformation, timeSince int) bool {

	newPod := p.Seen
	LastPod := lastSeen.Seen
	diff := newPod.Sub(LastPod)

	if diff > (time.Minute * time.Duration(timeSince)) {
		return true
	}

	return false

}
