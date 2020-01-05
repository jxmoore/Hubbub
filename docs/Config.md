# Configuration 
This document will go over the configuration options for Hubbub. 


## Overview
Hubub uses two forms of configuration, a json file passed in via the `-c` flag and enviroment variables. The enviroment variables are used in two cassees: 
1. If a field is nill but an associated enviroment variable has a value.
2. If the `-e` flag is used. 

At startup assuming the `-e` flag is not passed at runtime Hubbub will look for the file passed in via the `-c` flag, parse it and use that as the main source for its configuration; however, it also will attempt to populate any blank fields with enviroment variables (assuming those are set). This can be seen in <a href="bootstrap/bootstrap.go"> bootstrap.go</a> and <a href="models/config.go"> config.go</a> when bootstrap calls the LoadEnvVars method. So for example, assuming `-c 'file.json'` was passed in and this json file was missing the debug and the namespace it *would* be loaded from the enviroment variables if those variables were set :

```golang
	if !c.Debug && os.Getenv("HUBBUB_DEBUG") != "" {
		debug, err := strconv.ParseBool(os.Getenv("HUBBUB_DEBUG"))
		if err == nil {
			c.Debug = debug
		}
	}
	if c.Namespace == "" && os.Getenv("HUBBUB_NAMESAPCE") != "" {
		c.Namespace = os.Getenv("HUBBUB_NAMESAPCE")
	}
```

<br>
<br>

If the `-e` flag is set at runtime Hubbub ignores loading the file regardless if its passed in on the `-c` flag or not. 

```golang
	if !envOnly {
		if err := config.Load(path); err != nil {
			return fmt.Errorf("error loading config : \n%v", err)
		}
	}

	config.LoadEnvVars()
```


## Using a configuration file
The configuration file is fairly straightforward and this section will touch on its setup. To start lets take a look at the below json snippet which contains all of the configuration outside of the notifications :

```json
{
	"namespace": "The Namespace to watch",
	"Debug": true,
	"Self": "The name of the POD/deployment for Hubbub. Any pod containing this string will be excluded from notifications. ",
    "time": 5,
    "timezone": "America/New_York",
}
```

- **Namespace** : This is the Namespace in kubernetes that Hubbub will be watching
- **Debug** : This is to enable debug output. When debug is enabled Hubbub dumps extra output into STDOUT.
- **Self** : When a Pod change is detected Hubbub will exclude the change from notifications if it matches (fuzzy) *Self*, its used to prevent Hubbub from generating any noise if Hubbub encounters errors during rolling deployments; however, you could use it to exclude notifications from any pods really, *Self* just needs to be a part of the pod name.
- **Time** : This int represents a time in minutes that Hubbub should wait before alerting on a Pod if it has just encountered an error previously. This is somewhat dificult to get across so an example would likely be helpful. Assume a container *'x'* is deployed and it fails on startup generating an alert. Hubbub will save this instance as its 'LastSeen' pod. When Kubernetes restarts this pod in an attempt to get it running its going to fail again but Hubbub will see that the Pod matches what it already has in its 'LastSeen' so new notification will go out. **Unless** 'x' minutes (as specified in time) have passed. The default here is five.
- **TimeZone** : The timezone to convert times to, default is "America/New_York"

<br>

With those out of the way we can get to the Notifcations :

```json
{
    	"notifications": {
            "type": "slack",
            "slackWebhook": "your webhook",
            "slackChannel": "#kubeTroubles",
            "slackTitle":"The title of your slack post (default is defined in config.go)",
            "slackUser": "The user that this will post as (defaults to hubbub)",
            "slackIcon": "the users iscon for the post (default present in config.go)",
            "instrumentationKey" : "Your application insights instrumentation key",
            "customEventTitle" : "The title of the custom event that Hubbub will create in application insights", 
        },
}
```
We are not going to go deep into these as i feel they are fairly straightforward. But one thing worth mention is the `type` which represents the type of notification, the available options at this time are Slack and Application insights. If neither are used Hubbub will write the notifications as json to STDOUT.
 
> Also take note that if your using type "slack" you do not need the instrumentation key or custom event and the reverse can be said, no slack fields are needed if your type is application insights.

## ENV variables

There are a rather large number of enviorment variables that can be used, one for each configuration field found in the JSON. I beleive that most if not all of these should be self explanitory, this portion will list them and if neccessary have an excerpt about their use.

#### General : 
- **HUBBUB_DEBUG** : This is a *boolean*, so it should be 'true' or 'false'.
- **HUBBUB_NAMESAPCE**
- **HUBBUB_TIMECHECK** : This maps to the `time` field in the JSON. If this is abscent from the config and the env variable is nil Hubbub will default to 5.
- **HUBBUB_TIMEZONE**
- **HUBBUB_SELF** : If this is nil in the config and env variables 'Hubbub' will be used.

#### Slack specifics : 
- **HUBBUB_CHANNEL** 
- **HUBBUB_WEBHOOK**
- **HUBBUB_ICON**
- **HUBBUB_TITLE**

#### Application Insights :
- **HUBBUB_AIKEY** : The application insights instrumentation key.
- **HUBBUB_AITITLE** : The title of the application insights custom event. The default is `"There has been a pod error in production!"`.
