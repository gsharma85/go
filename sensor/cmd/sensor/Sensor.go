package main

import (
	. "github.com/gsharma85/go/sensor/internal/core"
	. "github.com/gsharma85/go/sensor/internal/data"
	"time"
	"flag"
)

func main() {
	
	dir := flag.String("d", "directory", "directory to watch")
	logfile := flag.String("lf", "log file", "where event logs will be writtern")
	flag.Parse()
	
	stopChan := make(chan struct{})
	keyGenFun := func(senseEvent SenseEvent) (groupByKey string, eventKey string) {
		return senseEvent.Action,senseEvent.Key
	}
	
	senseConfig := SenseConfig{stopChan, *dir, 100 * time.Millisecond, keyGenFun, *logfile}
	
	StartSensorPipeline("file", []string{"deduplicator"} , "filelog", senseConfig)
	
	<- stopChan
}
