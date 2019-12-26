package models

import (
	"strings"
	"testing"
	"time"
)

// Pod is a package wide PodStatusInformation used in all of the model tests as a base.
var TestPod = PodStatusInformation{
	Namespace:     "hubbub",
	PodName:       "hubbub",
	ContainerName: "hubbub",
	Image:         "hubbub",
	StartedAt:     time.Now(),
	FinishedAt:    time.Now(),
	ExitCode:      2,
	Reason:        "hubbub",
	Message:       "hubbub",
	Seen:          time.Now(),
}

// TestTimeCheck tests the timeCheck() method on the PodStatusInformation struct.
// This is also tested in the TestIsNew() function
func TestTimeCheck(t *testing.T) {

	testSuite := map[string]struct {
		expectedReturn bool
		timeDiff       time.Duration
		timeBack       int
	}{
		"timeCheck should return true #1": {
			expectedReturn: true,
			timeDiff:       time.Minute * -7,
			timeBack:       2,
		},
		"timeCheck should return true #2": {
			expectedReturn: true,
			timeDiff:       time.Hour * -1,
			timeBack:       15,
		},
		"timeCheck should return false #1": {
			expectedReturn: false,
			timeDiff:       time.Hour * 1,
			timeBack:       30,
		},
		"timeCheck should return false #2": {
			expectedReturn: false,
			timeDiff:       time.Minute * 50,
			timeBack:       25,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)
		fakePod := TestPod
		fakePod.Seen = time.Now().Add(testCase.timeDiff)
		ok := TestPod.timeCheck(fakePod, testCase.timeBack)

		if ok != testCase.expectedReturn {
			t.Errorf("expected %v but received %v", testCase.expectedReturn, ok)
		} else {
			t.Logf("received the expected response from timeCheck()")
		}
	}

}

// TestExitCode tests the ExitCodeLookup() on the PodStatusInformation struct
func TestExitCode(t *testing.T) {

	testSuite := map[string]struct {
		expectedReturn string
		exitCode       int
		PodInfo        PodStatusInformation
	}{
		"ExitCodeLookup should return Segmentation fault": {
			expectedReturn: "Segmentation fault.",
			exitCode:       139,
			PodInfo:        TestPod,
		},
		"ExitCodeLookup should return Application Error": {
			expectedReturn: "Application Error.",
			exitCode:       1,
			PodInfo:        TestPod,
		},
		"ExitCodeLookup should return Container terminated": {
			exitCode:       130,
			expectedReturn: "Container terminated.",
			PodInfo:        TestPod,
		},
		"ExitCodeLookup should return There was a error regardging...": {
			exitCode:       126,
			expectedReturn: "There was a error regardging permissions or the container could not be invoked.",
			PodInfo:        TestPod,
		},
		"ExitCodeLookup should return nil": {
			exitCode: 893,
			PodInfo:  TestPod,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)
		fakePod := TestPod
		fakePod.ExitCode = testCase.exitCode
		response := fakePod.ExitCodeLookup()

		if response != testCase.expectedReturn {
			t.Errorf("expected %v but received %v", testCase.expectedReturn, response)
		} else {
			t.Logf("received the expected response from IsNew()")
		}
	}

}

// TestConvertTime tests the ConvertTime() method on PodStatusInformation which converts time to EST.
func TestConvertTime(t *testing.T) {

	testSuite := map[string]struct {
		expectedZone string
		timeOffset   time.Duration
		PodInfo      PodStatusInformation
	}{
		"ConvertTime should EST #1": {
			expectedZone: "EST",
			PodInfo:      TestPod,
			timeOffset:   time.Minute * 19,
		},
		"ConvertTime should EST #2": {
			expectedZone: "EST",
			PodInfo:      TestPod,
			timeOffset:   time.Hour * 1,
		},
		"ConvertTime should EST #3": {
			expectedZone: "EST",
			PodInfo:      TestPod,
			timeOffset:   time.Hour * 8,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)

		timeZone := "America/New_York"
		fakePod := TestPod
		timeLocation, _ := time.LoadLocation(timeZone)

		fakePod.ConvertTime(timeLocation)
		dateArray := strings.Fields(fakePod.StartedAt.String())

		if dateArray[3] != testCase.expectedZone {
			t.Errorf("expected %v but received %v", testCase.expectedZone, dateArray[3])
		} else {
			t.Logf("received the expected response from ConvertTime()")
		}
	}

}

// TestIsNew tests the IsNew() method, which determines if a PodStatusInformation struct is new or not.
func TestIsNew(t *testing.T) {

	testSuite := map[string]struct {
		image            string
		continerName     string
		podName          string
		seen             time.Time
		finishedAt       time.Time
		startedAt        time.Time
		timeDif          time.Time
		timeBack         int
		exitCode         int
		expectedResponse bool
	}{
		"IsNew should return true": {
			image:            "hubbub",
			continerName:     "hubbub",
			podName:          "hubbub",
			expectedResponse: true,
			timeBack:         2,
		},
		"IsNew should return true due to the time difference": {
			image:            TestPod.Image,
			continerName:     TestPod.ContainerName,
			podName:          TestPod.PodName,
			seen:             time.Now().Add(time.Minute * -9),
			expectedResponse: true,
			timeBack:         2,
		},
		"IsNew should return false as pod and container names match": {
			image:            TestPod.Image,
			continerName:     TestPod.ContainerName,
			podName:          TestPod.PodName,
			finishedAt:       TestPod.FinishedAt,
			startedAt:        TestPod.StartedAt,
			seen:             TestPod.Seen,
			expectedResponse: false,
			timeBack:         20,
		},
		"IsNew should return false as the structs are identical": {
			expectedResponse: false,
			timeBack:         12,
		},
		"IsNew should return false because of nil values": {
			image:            "",
			continerName:     "",
			podName:          "",
			expectedResponse: false,
			timeBack:         12,
		},
		"IsNew should return false container name and exit code match": {
			image:            "hubbub",
			continerName:     TestPod.ContainerName,
			podName:          "hubbub",
			exitCode:         TestPod.ExitCode,
			finishedAt:       TestPod.FinishedAt,
			startedAt:        TestPod.StartedAt,
			seen:             TestPod.Seen,
			expectedResponse: false,
			timeBack:         21,
		},
	}

	for testName, testCase := range testSuite {

		t.Logf("\n\nRunning TestCase %v...\n\n", testName)
		fakePod := TestPod

		// everything will fall into this clause aside from the one testing deepequal
		// "IsNew should return false as the structs are identical"
		if testCase.podName != "" {
			fakePod.Image = testCase.image
			fakePod.ContainerName = testCase.continerName
			fakePod.PodName = testCase.podName
			fakePod.Seen = testCase.seen
			fakePod.StartedAt = testCase.startedAt
			fakePod.FinishedAt = testCase.finishedAt
		}
		timeZone := "America/New_York"
		timeLocation, _ := time.LoadLocation(timeZone)
		fakePod.ConvertTime(timeLocation)

		ok := TestPod.IsNew(fakePod, testCase.timeBack)
		if ok != testCase.expectedResponse {
			t.Errorf("expected %v but received %v", testCase.expectedResponse, ok)
		} else {
			t.Logf("received the expected response from IsNew()")
		}

	}

}
