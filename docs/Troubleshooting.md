
# Troubleshooting
This section will breifly go over some common issues and troubleshooting steps that can be taken. 

<br>
<br>

<details><summary><b>If you are receiving <i>Cannot create Pod event watcher....</i> in the console output</b></summary>
<ol>
  <li>Ensure the service account, role and rolebinding were created and reside within the correct namespace. </li>
  <li>Remember Hubbub requires <b>WATCH, GET and LIST</b> permissions.</li>
</ol>
</details>
<details><summary><b>If there are no messages being posted in slack channel <i>X</i></b></summary>
<ol>
  <li>Ensure your webhook is correct.</li>
  <li>Double check the channel if your supplying one in the config or env variables <i>channel</i> & <i>HUBBUB_CHANNEL</i> respectively.</li>
  <li>Check the console output. When Hubbub sends the notification if the response from slack is '<i>ok</i>' the console output should read '<i>Slack message sent...</i>', if the response body from slack does not match '<i>ok</i>' or there are errors the details will be written out as well.</li>
</ol>
</details>
<details><summary><b>If you think you are missing pod failure notifications</b></summary>
<ol>
  <li>Enable the Debug option in the config. This will print every change found on the channel to STDOUT.</li>
  <li>Check the config value for <i>TimeCheck</i> or env variable <i>HUBBUB_TIMECHECK</i> if one is supplied, this value represents a time in minutes and if too large could mean repetead or new failures are being deemed old.</li>
</ol>
</details>

<br>