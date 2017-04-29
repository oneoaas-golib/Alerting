package alert

import (
    "os"
    "fmt"
    "time"
    "strconv"
    "sync"

    "github.com/alerting/gocron"

    log "github.com/Sirupsen/logrus"
)

type AlertConfig struct {
    Alert_definition_path string  `json="Alert_definition_path"`
    Alert_definition_reload_timeinsec int64 `json="Alert_definition_reload_timeinsec"`
    Alert_query_endpoint string `json="Alert_query_endpoint"`
    Alert_snooze_path string `json="Alert_snooze_path"`
    Alert_log_path string `json="Alert_log_path"`
}


type AlertObject struct {
    status string
    now time.Time
    jobs []*gocron.Job
    md5 string

    AlertId string
    Enable bool
    Description string
    Cluster string
    Query string
    Operator string
    Threshold float64
    Frequency string
    AlertType string
    AlertValue string
}


type AlertList []*AlertObject

var AlertCfg AlertConfig 

var Logger *log.Entry

var AlertMap map[string]AlertList

var AlertSch *gocron.Scheduler

var AlertMutex sync.Mutex

var file *os.File

func init() {
   AlertSch = gocron.NewScheduler()
   AlertMap = make(map[string]AlertList)
}


func getLevel(i int) log.Level {
   switch i {
     case 1:
	return log.DebugLevel
     case 2:
	return log.InfoLevel
     case 3:
	return log.WarnLevel
     case 4:
	return log.ErrorLevel
     case 5:
	return log.FatalLevel
     default:
	fmt.Println("bad alert log level")
   }

   return log.WarnLevel
}


func LogInit(aLogLevel int) (err error) {
   // Log as JSON instead of the default ASCII formatter.
   log.SetFormatter(&log.JSONFormatter{})

   logPath := AlertCfg.Alert_log_path
   if logPath != "" {
      f, err := os.Stat(logPath)
      if err != nil {
         return err	
      }
      if !f.IsDir() {
	fmt.Errorf("log path provided must be a directory")
      }
      aMode := os.O_APPEND
      //TODO
      //if cfg.LogTruncate {
      //	aMode = os.O_TRUNC
      //}
      file, err = os.OpenFile(fmt.Sprintf("%s/alert.log", logPath), os.O_RDWR|os.O_CREATE|aMode, 0666)
      if err != nil {
	return err
      }
      //defer file.Close()
      log.SetOutput(file)
   } else {
      log.SetOutput(os.Stderr)
   }

   // Only log the warning severity or above.
   log.SetLevel(getLevel(aLogLevel))

   Logger = log.WithFields(log.Fields{
       "common": "Alerting Logging",
   })

   Logger.Info("Alert Logging....")

   return nil
}

func (a *AlertObject) Add() {
    a.status = "ADDED"
    a.now = time.Now()
}

func (a *AlertObject) UpdateMD5( md5 string) {
    a.md5 = md5
}

func (a *AlertObject) GetMD5()(string) {
     return a.md5 
}

func (a *AlertObject) UpdateStat(stat string) {
    if a.status == "RUNNING" && stat == "UPDATE" {
      for i := range a.jobs {
         //Remove job from scheduler
         AlertSch.RemoveJob(a.jobs[i])
      } 
      a.jobs = nil
    }     
    
    //update status
    a.status = stat 
}


func (a *AlertObject) GetStatus()(string) {
    return a.status
}

func (a *AlertObject) NewJobs(){
  a.jobs = make([]*gocron.Job,0,1)
}

func (a *AlertObject) AddJobs( freq string, val string, fn interface{}){
   var tJob *gocron.Job
   switch freq{
   case "SEC":
     interval,_ := strconv.Atoi(val)
     tJob = AlertSch.Every(uint64(interval)).Seconds()
     tJob.Do(fn,"SEC",a)
   case "MIN":
     interval,_ := strconv.Atoi(val)
     tJob = AlertSch.Every(uint64(interval)).Minutes()
     tJob.Do(fn,"MIN",a) 
   case "HOURLY":
     interval,_ := strconv.Atoi(val)
     tJob = AlertSch.Every(uint64(interval)).Hours()
     tJob.Do(fn,"HOURLY",a)
   case "DAILY":
     tJob = AlertSch.Every(1).Day().At(val)
     tJob.Do(fn,"DAILY",a)
   case "MONDAY":
     tJob = AlertSch.Every(1).Monday().At(val)
     tJob.Do(fn,"MONDAY",a)
   case "TUESDAY":
     tJob = AlertSch.Every(1).Tuesday().At(val)
     tJob.Do(fn,"TUESDAY",a)
   case "WEDNESDAY":
     tJob = AlertSch.Every(1).Wednesday().At(val)
     tJob.Do(fn,"WEDNESDAY",a)
   case "THURSDAY":
     tJob = AlertSch.Every(1).Thursday().At(val)
     tJob.Do(fn,"THURSDAY",a)
   case "FRIDAY":
     tJob = AlertSch.Every(1).Friday().At(val)
     tJob.Do(fn,"FRIDAY",a)
   case "SATURDAY":
     tJob = AlertSch.Every(1).Saturday().At(val)
     tJob.Do(fn,"SATURDAY",a)
   case "SUNDAY":
     tJob = AlertSch.Every(1).Sunday().At(val)
     tJob.Do(fn,"SUNDAY",a)
   default:
     Logger.Error(fmt.Sprintln("Alert frequency is not supported =", freq))
   }
   a.jobs = append(a.jobs,tJob)
}

func DeleteJobs(fn string, sIdx int) {
   Logger.Info("Inside Delete Job")

   AlertMutex.Lock() 
   al := AlertMap[fn]
   Logger.Info(fmt.Sprintln("Before delete Len of AlertObj =",len(al)))
   for i:=sIdx; i < len(al); i++ {
    for j := range al[i].jobs {
      //Remove job from scheduler
      AlertSch.RemoveJob(al[i].jobs[j])
    }
    al[i].jobs = nil
    tmp := append(al[:i],al[i+1:]...)
    al = nil
    al = tmp
   }
   Logger.Info(fmt.Sprintln("After delete Len of AlertObj =",len(al)))
   AlertMap[fn] = al
   AlertMutex.Unlock()
}

func DeleteAlertObj(fn string, idx int) {
  Logger.Info("Inside Delete AlertObj")

  al := AlertMap[fn]

  tmp := append(al[:idx],al[idx+1:]...)

  if len(tmp) > 0 {
    AlertMutex.Lock()
    AlertMap[fn] = tmp
    AlertMutex.Unlock()
  } else {
    tmp = nil
    AlertMutex.Lock()
    AlertMap[fn] = nil
    AlertMutex.Unlock()
  }

  al = nil
}
