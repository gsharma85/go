package out

import (
	. "github.com/gsharma85/go/sensor/internal/data"
	"log"
	"os"
)

type LogToFile struct {
	StopSignal chan struct{}
	fileName string
}


func (logToFile LogToFile) Deposit(inChan chan SenseEvent) {
	go func() {
		closeRoutine := false
		logFile, err := os.OpenFile(logToFile.fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
		if err != nil {
		   panic(err)
		}
		logger := log.New(logFile, "", log.LstdFlags)
		for {
			select {
				case event, _ := <- inChan:
				logger.Println("Got event: %s", event)
			
				case _, open := <- logToFile.StopSignal:
				if !open {
					closeRoutine = true
				}
			}
			if closeRoutine {
				return
			}
		}
	} ()
}

func CreateFileLogDepositor(senseConfig SenseConfig) (LogToFile,bool) {
	return LogToFile{senseConfig.StopSignal, senseConfig.Logfile},true
}