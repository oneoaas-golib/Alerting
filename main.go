package main

import (
    "fmt"
    "flag"
    "time"
    "os"
    "encoding/json"

    "github.com/alerting/alert"
    "github.com/alerting/input"
    "github.com/alerting/scheduler"
)


var AlertLogLevel = flag.Int("log-level",3,"1: debug, 2: info, 3: warning, 4: error, 5: fatal")

var ConfigFile = flag.String("config-file", "", "specify the json config file from which to load the configuration.")



func loadConfig() (err error) {
   flag.Parse()
   if *ConfigFile == "" || *AlertLogLevel <= 0 || *AlertLogLevel > 5 {
      flag.Usage()
      return fmt.Errorf("Error command line  config")
   }

   file,err := os.Open(*ConfigFile)
   if err != nil {
      fmt.Println("Can't Open Alert Config File")
   }
   defer file.Close()
   decoder := json.NewDecoder(file)
   err = decoder.Decode(&alert.AlertCfg)
   if err != nil {
      fmt.Println("Alert Config Load Error =", err)
      return err
   } 

   //fmt.Println(alert.AlertCfg.Alert_definition_path)
   //fmt.Println(alert.AlertCfg.Alert_definition_reload_timeinsec)
   //fmt.Println(alert.AlertCfg.Alert_query_endpoint)
   //fmt.Println(alert.AlertCfg.Alert_snooze_path)
   //fmt.Println(alert.AlertCfg.Alert_log_path)

   return nil 
}


func main() {

   fmt.Println("Inside Main")

   err := loadConfig()
   if err != nil {
      fmt.Println("Error in Loading Config File = ", err)
      return
   }

   err = alert.LogInit(*AlertLogLevel) 
   if err != nil {
      fmt.Println("Error in Logging init = ", err)
      return
   }

   //Start Alert WebServer
   go input.AlertWebServer()

   //Start Alert Scheduler
   go scheduler.SchStart() 

   for {
     input.CheckForUpdate()
     
     time.Sleep(time.Second * time.Duration(alert.AlertCfg.Alert_definition_reload_timeinsec))
   }

   fmt.Println("End Main")
}
