package output

import(
   "fmt"
   "encoding/json"
   "io/ioutil"
   "net/http"
   "net/url"
   "time"
   "errors"
   "strings"
   "strconv"
   "os"
   "bufio"

   "github.com/alerting/alert"
)

type tagsetType struct {
   Cluster string `json:"cluster,omitempty"`
   Host string `json:"host,omitempty"`
}

type scalarType struct {
   Tagset tagsetType `json:"tagset,omitempty"`   
   Value float64 `json:"value,omitempty"`
}

type seriesType struct {
   Tagset tagsetType `json:"tagset,omitempty"`
   Values []float64 `json:"values,omitempty"`
}

type bodyType struct {
   Query string `json:"query"`
   Name string `json:"name"`
   Type string `json:"type,omitempty"`
   Scalars []scalarType `json:"scalars,omitempty"`
   Series []seriesType `josn:"series,omitempty"`
}

type respType struct {
   Success bool `json:"success"`
   Message string `json:"message,omitempty"`
   Body []bodyType `josn:"body,omitempty"`
}

type currData struct {
   Tagset tagsetType 
   Value float64 
}

type alertResp struct {
   Query string
   CurrentValue []currData
   Threshold float64
}


func send(aAlertType, aTo , aSub, aBody string) {

   alert.Logger.Info(fmt.Sprintln("Alert Type =", aAlertType))
   switch aAlertType {
   case "Mail":
      SendMail(aBody, aTo, aSub)
   default:
      alert.Logger.Info(fmt.Sprintf("Alert Type is not Set"))
   }
}

func postQuery(query string) (res respType,  err error) {

    var Url *url.URL
    Url, err = url.Parse(alert.AlertCfg.Alert_query_endpoint)
    if err != nil {
       return res,err
    }

    parameters := url.Values{}
    parameters.Add("query", query)
    Url.RawQuery = parameters.Encode()

    req, err := http.NewRequest("GET", Url.String(), nil)

    alert.Logger.Debug(fmt.Sprintf("req =",req))

    timeout := time.Duration(2 * time.Second)

    client := &http.Client{
                Timeout: timeout,
              }
    resp, err := client.Do(req)

    alert.Logger.Debug(fmt.Sprintf("resp =",resp,"err =", err))

    if err != nil {
       return res, err
    }
    defer resp.Body.Close()

    alert.Logger.Debug(fmt.Sprintf("response Status:", resp.Status))
    alert.Logger.Debug(fmt.Sprintf("response Headers:", resp.Header))
    rBody, err := ioutil.ReadAll(resp.Body)
    alert.Logger.Debug(fmt.Sprintf("response Body:", string(rBody)))

    err = json.Unmarshal(rBody,&res)
    if err != nil {
        return res,err
    }
    if res.Success == false {
       return res, errors.New(res.Message)
    }

    alert.Logger.Debug(fmt.Sprintf("Json Rsponce = ",res))

    return res,nil
}

func checknConfirm( operator string, threshold float64, currData float64) bool {

    switch operator {

    case ">":
      if currData > threshold {
        return true
      } 

    case "<":
      if currData < threshold {
        return true
      } 

    case "=":
      if currData == threshold {
        return true
      }
    default:
         alert.Logger.Info(fmt.Sprintf("Operate not supported ", operator))
    }//switch

    return false
}

func readFromFile(fn string)(string, error) {
  f, err := os.Open(fn)
  if err != nil {
    alert.Logger.Error(fmt.Sprintf("error opening file ", err))
    return "", err
  }
  defer f.Close()
  r := bufio.NewReader(f)
  return r.ReadString('\n')
}

func checkForSnooze(aID string, cluster string, host string) bool {
   alert.Logger.Debug(fmt.Sprintf("AlertID =", aID, "cluster =", cluster, "host =", host))

   //Read list of files
   files, err := ioutil.ReadDir(alert.AlertCfg.Alert_snooze_path + "/" + aID)
   if err != nil {
      alert.Logger.Error(fmt.Sprintf(" Read Dir Error =", err))
      return false
   }

   for _,file := range files {
     var val []byte
     if file.Name() ==  cluster {
        val,err = ioutil.ReadFile(alert.AlertCfg.Alert_snooze_path + "/" + aID + "/" + cluster)
        if err != nil {
          alert.Logger.Error(fmt.Sprintf(" Read Error =", err))
          continue
        }    
     } else if file.Name() ==  host {
        val,err = ioutil.ReadFile(alert.AlertCfg.Alert_snooze_path + "/" + aID + "/" + host)
        if err != nil {
          alert.Logger.Error(fmt.Sprintf(" Read Error =", err))
          continue
        }
     }
     substr := strings.Split(string(val)," ")
     var v float64
     if v,err = strconv.ParseFloat(substr[0],64); err != nil {
        continue 
     }
        
     delta := time.Now().Sub(file.ModTime())
     alert.Logger.Debug(fmt.Sprintf("Given duration =",v," current delta =",delta.Hours()))
     if v > delta.Hours(){
        return true
     }
   }//for
   return false
}

func checknAlert (ao *alert.AlertObject) {

    var tar alertResp
    tcd := make([]currData,0,1)
    sure := false

    res, err := postQuery(ao.Query)
    if err != nil {
      alert.Logger.Error(fmt.Sprintf("postQuery err =",err))
      return
    }

    if res.Body[0].Type == "series" {
      s := res.Body[0].Series
      for i:=0; i < len(res.Body[0].Series); i++ {
         for j:=0; j < len(s[i].Values); j++ {
            if checknConfirm(ao.Operator, ao.Threshold, s[i].Values[j]) == true {
               if checkForSnooze(ao.AlertId, s[i].Tagset.Cluster, s[i].Tagset.Host) == false {
                  sure = true
                  var t currData
                  t.Tagset.Cluster = s[i].Tagset.Cluster
                  t.Tagset.Host = s[i].Tagset.Host
                  t.Value = s[i].Values[j]

                  tcd = append(tcd,t)
               }//check for snooze
            }//check n confirm
         }// for values
      } // for series
    } else if res.Body[0].Type == "scalars" {
      s := res.Body[0].Scalars
      for i:=0; i < len(res.Body[0].Scalars); i++ {
         if checknConfirm(ao.Operator, ao.Threshold, s[i].Value) == true {
           if checkForSnooze(ao.AlertId, s[i].Tagset.Cluster, s[i].Tagset.Host) == false {
              sure = true
              var t currData
              t.Tagset.Cluster = s[i].Tagset.Cluster
              t.Tagset.Host = s[i].Tagset.Host
              t.Value = s[i].Value

              tcd = append(tcd,t)
           }//check for snooze
         } //check n confirm
      }// for

    }//scalar

    if sure == true {
       tar.Query = ao.Query
       tar.CurrentValue = tcd
       tar.Threshold = ao.Threshold

       //jb,_ := json.Marshal(tar)
       jb,_ := json.MarshalIndent(tar,"", "  ")

       alert.Logger.Info("Alerting............")
       msg := fmt.Sprintf("%v", string(jb))
       send(ao.AlertType, ao.AlertValue, ao.Description, msg)
    } else {
      tcd = nil
    }
}

func  AlertTask( freq string, ao *alert.AlertObject) {
    alert.Logger.Debug(fmt.Sprintf("Inside AlterTask  freq = ",freq,"alertObj =",ao))

    checknAlert(ao)
}
