package out

import (
	. "github.com/gsharma85/go/sensor/internal/data"
	"log"
)

type LogToConsole struct {
	StopSignal chan struct{}
}

func (logger LogToConsole) Deposit(inChan chan SenseEvent) {
	go func() {
		closeRoutine := false
		for {
			select {
				case event, _ := <- inChan:
				log.Println("Got event: %s", event)
			
				case _, open := <- logger.StopSignal:
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

func CreateConsoleDepositor(senseConfig SenseConfig) (LogToConsole,bool) {
	return LogToConsole{senseConfig.StopSignal},true
}

