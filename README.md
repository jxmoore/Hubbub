# Hubbub
HubBub produces alerts for Pod and Container failures.

Hubbub is a small application that runs locally inside of a Kubernetes cluster with a given service account, alerting on pod and container failures. It has no external dependencies outside of the STDLIB and Kubernetes packages.

Download:
```shell
go get github.com/jxmoore/hubbub
```
If you do not have the go command on your system, you need to [Install Go](http://golang.org/doc/install) first.

* * *

Hubbub runs indeffinatly and listens for pod and container changes on a channel until the application is terminated. If a pod failure is seen on the channel the information about the pod and container is sent as a notification using STDOUT or Slack.

<p align="center">
  <img width="590" height="165"  src="images/segfault.PNG">
  <img width="590" height="165"  src="images/eviction.PNG">
</p>

> When a change is seen Hubbub checks to see if its new, this is done using a very soft match.
> This is to prevent noise, for example to ensure a constant restarting container does not generate constant... hubbub. If the same  
> container/pod was last seen five minutes ago the notification is considered new again and will be sent. 
	
>> See the IsNew() method on the PodStatusInformation struct in kube.go for information on how the decision is made.

## Build
If building the application locally do so outside of $GOPATH and ensure that you are using at least go V1.12. Aside from that a simple `go build .` in the PWD should result in a built binary. However Hubbub uses the InCluster config, so a built locally binary will do you no good unless your troubleshooting the build or planning on moving it to a container in a cluster. The supplied dockerfile can be used to build a useable docker image.

## Usage
Update the yaml file supplied to create a service account, cluster role binding etc... ensure that all reside within the correct namespace and the image matches either the newest image for Hubbub or one you have built. If you have built this image yourself and supplied your own config you should be good to go.

If you have *NOT* supplied a new config however and are using the Hubbub image you can :
- Create a config, present it to the container as a mount and point to it via -c in the CMD
- Use the -e flag and specify all settings as env variables, see `config.go` for more information on the enviroment variables used..

## Errors 
- If you are receiving `Cannot create Pod event watcher....` in the console output ensure the service account, role and rolebinding are correct. It requires WATCH, GET and LIST permissions.
- If there are no messages being posted in slack channel 'x' ensure your webhook is correct. Assuming it is be sure to check the console output. When sending the notification if no error is received the console should read 'Slack message sent...' along with the response body from slack, if there are errors these are written out as well.
- If you think you are missing changes check the before mentioned IsNew method, it could be that the changes you expect are being deemed 'old'.

### TODO 
- Tests.
- Parse the slack response body to ensure 'ok' is received back.
- Implement the LabelSelectors.
- ~The IsNew() should do a check on duration and ensure if 'x' time has passed the notification should be sent regardless.~
- The 'value' field in the slack post should be exsposed in the config and should take Go templating syntax.
- Fancy up readme, include config example
