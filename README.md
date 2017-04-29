
# Alerting System in Go Lang 
Alerting system, execute alert query with Metric Square Engine(MQE), which retrun result. Result value compared with threashold value, if it satisfies, alert will be sent based on alert type. Gocron scheduler is used to schedule alert defintion based alert config.
---

### for Building and Running

Build and install 
 1) Move to srouce path $GOPATH/src/github/alert
 2) go install ./...

Run
 nohup $GOPATH/bin/alert &

####### Alert Definition File#######

 Alert Definition file contains all alert define for the team. All the fields are mandatory. And its defined in json format
under config/alert/alert.json

####### Alert Snooze #######
   Snooze can be set to each alert Id. Alert ID is generated based on file name and Alert ID field in the file. Each alert ID will have Cluster and Hosts. Based on cluster or hosts the snooze can be set. Snooze will set on hourly basis.

   Here is the Ui using which snooze can be set and unset for each AlertId. http://localhost:9999/alert-snooze-ui

