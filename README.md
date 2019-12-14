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

If you have **NOT** supplied a new config however and are using the Hubbub image you can :
- Create a config, present it to the container as a mount and point to it via -c in the CMD
- Use the -e flag and specify all settings as env variables, see `config.go` for more information on the enviroment variables used.

## Errors 
- If you are receiving `Cannot create Pod event watcher....` in the console output : 
  1. Ensure the service account, role and rolebinding were created and reside within the correct namespace. 
  2. Remember Hubbub requires **WATCH, GET and LIST** permissions.
- If there are no messages being posted in slack channel *X* : 
  1. Ensure your webhook is correct.
  2. Double check the channel if your supplying one in the config or env variables `channel` & `HUBBUB_CHANNEL` respectively. 
  2. Check the console output. When Hubbub sends the notification if the response from slack is '*ok*' the console output should read '*Slack message sent...*', if the response body from slack does not match '*ok*' or there are errors the details will be written out as well.
- If you think you are missing pod failure notifications: 
  1. Enable the Debug option in the config. This will print every change found on the channel to STDOUT. 
  2. Check the config value for `TimeCheck` or env variable `HUBBUB_TIMECHECK` if one is supplied, this value represents a time in minutes and if too large could mean repetead or new failures are being deemed old.
  3. Check the before mentioned IsNew method, it could be that the changes you expect are being deemed 'old'.

### TODO 
- Tests. -- In Progress
- Fancy up readme, include config example. -- In Progress
- Update GetKubeClient() to allow using a local config for testing outside of a cluster.
- Update comments - many of which are no longer applicable as the code has changed.
- Implement the LabelSelectors.
- The 'value' field in the slack post should be exsposed in the config and should take Go templating syntax.
- ~The IsNew() should do a check on duration and ensure if 'x' time has passed the notification should be sent regardless.~
- ~Parse the slack response body to ensure 'ok' is received back.~