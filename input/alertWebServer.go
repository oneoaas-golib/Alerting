package input

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"

	"github.com/alerting/alert"
)

type alertIdType struct {
	AlertId []string `json:"AlertId"`
}

const HTML = `
 <!DOCTYPE html>
 <html lang="en">
    <head>
      <meta charset="utf-8">
      <title> RTM Snooze </title>
    </head>
 <body>
     <p>
     <form action="/alert-snooze-ui" method="POST">
     <label>AlertID:</label>
     <select name="AlertID"> 
       {{range $key, $value := .}}
          <option value="{{ $value }}">{{ $key }}</option>
       {{end}}
     </select>
        Cluster Or Host Name :<input type="text" name="Cluster_Or_Host">
        Value in Hours :<input type="number" name="Value" min="0" max="48">
        <input type="submit" value="Save">
     </form>
     </p>
 </body>
 </html>`

func getAlertIds() map[string]interface{} {
	tmp := make(map[string]interface{})

	alert.AlertMutex.Lock()
	for k, _ := range alert.AlertMap {
		al := alert.AlertMap[k]
		for j := 0; j < len(al); j++ {
			tmp[al[j].AlertId] = al[j].AlertId
		}
	}
	alert.AlertMutex.Unlock()
	alert.Logger.Warn("AlertId =", tmp)

	return tmp
}

func writeToFile(dir, fn, val string) {
	alert.Logger.Warn(fmt.Sprintf(dir, fn, val))

	tmp, _ := strconv.Atoi(val)
	if tmp < 0 {
		alert.Logger.Info(fmt.Sprintf("Value can't be in negative, So reset to default min value to 0"))
		val = "0"
	}

	if tmp > 48 {
		alert.Logger.Info(fmt.Sprintf("Value can't be higher than two days, So reset to default max value to 48"))
		val = "48"
	}

	if _, err := os.Stat(alert.AlertCfg.Alert_snooze_path + "/" + dir); os.IsNotExist(err) {
		os.MkdirAll(alert.AlertCfg.Alert_snooze_path+"/"+dir, 0777)
	}

	ioutil.WriteFile(alert.AlertCfg.Alert_snooze_path+"/"+dir+"/"+fn, []byte(val), 0644)
}

func alertSnoozeUi(w http.ResponseWriter, r *http.Request) {

	alertArr := getAlertIds()
	if len(alertArr) == 0 {
		alertArr = nil
		return
	}

	ddTemplate, err := template.New("Alert").Parse(string(HTML))
	if err != nil {
		alert.Logger.Error(fmt.Sprintf("html template err =", err))
		return
	}

	if r.Method == "POST" {
		r.ParseForm()

		alert.Logger.Warn(fmt.Sprintf("Alert ID", r.Form["AlertID"][0]))
		alert.Logger.Warn(fmt.Sprintf("Cluster or Host", r.Form["Cluster_Or_Host"][0]))
		alert.Logger.Warn(fmt.Sprintf("Value", r.Form["Value"][0]))
		if r.Form["AlertID"][0] == "" {
			return
		}
		if r.Form["Cluster_Or_Host"][0] == "" {
			return
		}
		if r.Form["Value"][0] == "" {
			return
		}

		writeToFile(r.Form["AlertID"][0], r.Form["Cluster_Or_Host"][0], r.Form["Value"][0])
	} else {
		ddTemplate.Execute(w, alertArr)
	}
}

func getAlertId(w http.ResponseWriter, r *http.Request) {
	var tmp alertIdType
	tmp.AlertId = make([]string, 0, 1)

	alert.AlertMutex.Lock()
	for k, _ := range alert.AlertMap {
		al := alert.AlertMap[k]
		for j := 0; j < len(al); j++ {
			tmp.AlertId = append(tmp.AlertId, al[j].AlertId)
		}
	}
	alert.AlertMutex.Unlock()
	alert.Logger.Info(fmt.Sprintf("AlertId =", tmp))

	b, _ := json.Marshal(tmp)

	w.Write(b)

}

func AlertWebServer() {
	http.HandleFunc("/GetAlertID", getAlertId)
	http.HandleFunc("/alert-snooze-ui", alertSnoozeUi)
	err := http.ListenAndServe(":9999", nil)
	if err != nil {
		alert.Logger.Fatal("alertWebServer: ListenAndServe: ", err)
	}
}
