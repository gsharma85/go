package main

import (
	. "github.com/gsharma85/go/sensor/internal/core"
	. "github.com/gsharma85/go/sensor/internal/data"
	"time"
)

func main() {
	stopChan := make(chan struct{})
	keyGenFun := func(senseEvent SenseEvent) (groupByKey string, eventKey string) {
		return senseEvent.Action,senseEvent.Key
	}
	
	senseConfig := SenseConfig{stopChan, "C:\\gotest", 100 * time.Millisecond, keyGenFun}
	
	StartSensorPipeline("file", []string{"deduplicator"} , "console", senseConfig)
	
	<- stopChan
}
