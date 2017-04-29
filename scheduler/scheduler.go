package scheduler

import(
    "fmt"

    "github.com/alerting/alert"
)

type scheduler struct {
  status string
  stopChan chan bool
}


var sch scheduler


func SchStart() {
   alert.Logger.Debug("Inside Sch starting")
   sch.status = "STARTED"
   sch.stopChan =  alert.AlertSch.Start()
  
   Running()
}

func Running() {
   alert.Logger.Debug("Inside Sch Running")
   for {
      switch {
      case <-sch.stopChan:
          close(sch.stopChan) 
          sch.stopChan = nil
      }
   }
}

func SchStop() {
   sch.stopChan <- true
}


func Inform() {
   alert.Logger.Debug("Inside Schedule Inform")
   alert.AlertMutex.Lock()
   for k,_ := range alert.AlertMap {
      al := alert.AlertMap[k]
      alert.Logger.Info(fmt.Sprintln("Len of alertob listj =",len(al)))
      for i:=0; i < len(al); i++ {
        ao := al[i]
        if ao.GetStatus() == "ADDED" || ao.GetStatus() == "UPDATE" {
           ao.UpdateStat("RUNNING") 
        }
      }
   }
   alert.AlertMutex.Unlock()

   //Run list of jobs
   alert.AlertSch.RunPending()
}
