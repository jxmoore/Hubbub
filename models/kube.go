package models

import (
	"fmt"
	"reflect"
	"strconv"
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

// Load takes a *v1.Pod and loads the attributes into the PodStatusInformation struct. it has some
// logic to ensure the we dont alert on pending/succefully completed pods and things of that nature.
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
// that a instance of the struct is new and not a repeat or close enough to be considered a repeat.
//
// This is done to cut down on redundant messages generated in Watch(). For e.g. a pod that terminates
// shortly after startup due to an error in program.cs will be restarted
// by kubernetes where it will proceed to fail again and constantly generate alerts as it goes up and down.
//
// Is new returns true if the pod has not been seen for 'x' minutes or it does not match any of the criteria
func (p PodStatusInformation) IsNew(lastSeen PodStatusInformation, timeSince int) bool {

	// assume not a failure
	if p.Image == "" || p.ContainerName == "" || p.FinishedAt.IsZero() {
		return false
	}

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

// ConvertTime converts all of the times found in p to local (EST). This is in place because some
// users host their Kubeernetes clusters in cloud enviroments where the local timezone does not match
// the end users.
func (p *PodStatusInformation) ConvertTime(tlocal *time.Location) {

	p.FinishedAt = p.FinishedAt.In(tlocal)
	p.StartedAt = p.StartedAt.In(tlocal)

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

// podErrorReason returns a string composed of p.Reason and/or p.Message depending upon
// their values. If both are nil a user friendly nil is returned.
func podErrorReason(p PodStatusInformation) string {

	if p.Reason != "" && p.Message != "" {
		return fmt.Sprintf("Failure reason received : `%v - %v`", p.Reason, p.Message)
	} else if p.Message != "" {
		return fmt.Sprintf("Failure reason received : `%v`", p.Message)
	} else if p.Reason != "" {
		return fmt.Sprintf("Failure reason received : `%v`", p.Reason)
	} else {
		return "Unable to determine the reason for the failure."
	}

}

// podErroCode returns a concat of the errorcode (int) and that error codes meaning if one is returned via ExitCodeLookup()
func podErrorCode(p PodStatusInformation) string {

	errorDetails := strconv.Itoa(p.ExitCode)
	errInfo := p.ExitCodeLookup()

	if errInfo != "" {
		return fmt.Sprintf("Error code : %v `%v`", errorDetails, errInfo)
	}

	return fmt.Sprintf("Error code : %v\n", errorDetails)

}
