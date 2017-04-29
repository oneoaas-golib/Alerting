package input

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"time"

	"github.com/alerting/alert"
	"github.com/alerting/output"
	"github.com/alerting/scheduler"
)

type param struct {
	Alert_id         string  `json:"alert_id"`
	Enable           bool    `json:"enable"`
	Description      string  `json:description"`
	Frequency        string  `json:"frequency"`
	Query            string  `json:"query"`
	Operator         string  `json:"operator"`
	Threshold        float64 `json:"threshold"`
	Exclude_def_team bool    `json:"exclude_def_team"`
	Additional_team  string  `json:"additional_team"`
}

type alertStruct struct {
	Alert_type  string  `json:"alert_type"`
	Def_team    string  `json:"def_team"`
	Def_cluster string  `json:"def_cluster"`
	Params      []param `json:"params"`
}

var input struct {
	fStats map[string]time.Time
}

func init() {
	input.fStats = make(map[string]time.Time)
}

func getFrequency(obj *alert.AlertObject, freq string) {

	sp := strings.Split(freq, ";")

	for i := 0; i < len(sp); i++ {
		name := strings.Split(sp[i], "(")
		alert.Logger.Debug(fmt.Sprintf("name =", name[0]))
		val := strings.Split(name[1], ")")
		alert.Logger.Debug(fmt.Sprintf("val =", val))
		values := strings.Split(val[0], ",")
		for j := 0; j < len(values); j++ {
			obj.AddJobs(name[0], values[j], output.AlertTask)
		} //for j
	} //for i
}

func getMD5Hash(b []byte) string {
	hasher := md5.New()
	hasher.Write(b)
	return hex.EncodeToString(hasher.Sum(nil))
}

func UpdateNow(co *alert.AlertObject, no *alert.AlertObject) {
	alert.Logger.Debug(fmt.Sprintf("Inside Update Now co =", co, " new =", no))

	co.UpdateStat("UPDATE")

	co.AlertId = no.AlertId
	co.Enable = no.Enable
	co.Description = no.Description
	co.Cluster = no.Cluster
	co.Query = no.Query
	co.Operator = no.Operator
	co.Threshold = no.Threshold
	co.Frequency = no.Frequency
	co.AlertType = no.AlertType
	co.AlertValue = no.AlertValue
	co.UpdateMD5(no.GetMD5())

	// Enable is false don't add to active job list
	if co.Enable == false {
		return
	}

	//add jobs based on frquency requested
	if co.Frequency != "NULL" {
		co.NewJobs()
		getFrequency(co, co.Frequency)
	}
}

func ReadFromFile(fp string, fn string) {

	alert.Logger.Debug(fmt.Sprintf("Inside ReadFromFile fn =", fn))

	byt, err := ioutil.ReadFile(fp + "/" + fn)
	if err != nil {
		alert.Logger.Error(fmt.Sprintf("Error in reading file =", err))
		return
	}

	alert.Logger.Debug(fmt.Sprintf("input string =", string(byt)))

	var data alertStruct

	err = json.Unmarshal(byt, &data)
	if err != nil {
		alert.Logger.Error(fmt.Println("Config Unmarshal error =", err))
	}

	alert.Logger.Debug(fmt.Sprintf("Unmarshal =", data))

	tAlertList := alert.AlertMap[fn]
	tlen := len(tAlertList)

	//Delete additional jobs
	if tlen > len(data.Params) {
		alert.DeleteJobs(fn, len(data.Params))
		tAlertList = alert.AlertMap[fn]
		tlen = len(tAlertList)
	}

	for i := 0; i < len(data.Params); i++ {
		tmp := data.Params[i]
		obj := new(alert.AlertObject)

		sf := strings.Split(fn, ".")

		obj.AlertId = sf[0] + "-" + tmp.Alert_id
		obj.Enable = tmp.Enable
		obj.Description = tmp.Description
		obj.Query = tmp.Query
		obj.Cluster = data.Def_cluster
		obj.Operator = tmp.Operator
		obj.Threshold = tmp.Threshold
		obj.Frequency = tmp.Frequency
		obj.AlertType = data.Alert_type
		obj.AlertValue = data.Def_team

		if tmp.Exclude_def_team == true && tmp.Additional_team != "NULL" {
			obj.AlertValue = tmp.Additional_team
		}

		if tmp.Exclude_def_team == false && tmp.Additional_team != "NULL" {
			obj.AlertValue = obj.AlertValue + tmp.Additional_team
		}

		b, err := json.Marshal(obj)
		if err != nil {
			alert.Logger.Error(fmt.Sprintf("Cannot Unmarshal =", err))
		}

		obj.UpdateMD5(getMD5Hash(b))

		//For New data
		if tlen == 0 || i >= tlen {

			//Alert Enable is false add to alert obj list
			//But active jobs will not be created
			if obj.Enable == false {
				//add
				alert.AlertMutex.Lock()
				obj.Add()
				tAlertList = append(tAlertList, obj)
				alert.AlertMap[fn] = tAlertList
				alert.AlertMutex.Unlock()
				continue
			} else { //enable == true

				//add active jobs based on frquency requested
				if obj.Frequency != "NULL" {
					obj.NewJobs()
					getFrequency(obj, obj.Frequency)
				}

				//add
				alert.AlertMutex.Lock()
				obj.Add()
				tAlertList = append(tAlertList, obj)
				alert.AlertMap[fn] = tAlertList
				alert.AlertMutex.Unlock()
			} //enable == true

		} else { //Already exists

			//already exists
			tObj := tAlertList[i]

			//check for Update
			if tObj.GetMD5() == obj.GetMD5() {
				alert.Logger.Debug(fmt.Sprintf("No Udpate =", tObj))
				continue
			} else {
				//Found update
				alert.Logger.Debug(fmt.Sprintf("!!!!!!!!Udpate "))
				UpdateNow(tObj, obj)

			} //if == md5

		} //if already exist data

	} //End of For
}

func CheckForUpdate() {

	alert.Logger.Debug("Inside Check for Update")

	//Read list of files
	files, err := ioutil.ReadDir(alert.AlertCfg.Alert_definition_path)
	if err != nil {
		alert.Logger.Error(fmt.Sprintf(" Read Dir Error =", err))
	}

	for _, file := range files {

		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		if v, ok := input.fStats[file.Name()]; ok {
			if v == file.ModTime() {
				continue
			}
		}

		if _, ok := alert.AlertMap[file.Name()]; !ok {
			tmp := make([]*alert.AlertObject, 0, 1)
			alert.AlertMutex.Lock()
			alert.AlertMap[file.Name()] = tmp
			alert.AlertMutex.Unlock()
		}

		input.fStats[file.Name()] = file.ModTime()
		ReadFromFile(alert.AlertCfg.Alert_definition_path, file.Name())
		scheduler.Inform()

	} //for
}
