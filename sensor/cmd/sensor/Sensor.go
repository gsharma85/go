package main

import (
	. "github.com/gsharma85/go/sensor/internal/core"
	. "github.com/gsharma85/go/sensor/internal/data"
	"time"
	"flag"
)

func main() {
	
	dir := flag.String("d", "directory", "directory to watch")
	flag.Parse()
	
	stopChan := make(chan struct{})
	keyGenFun := func(senseEvent SenseEvent) (groupByKey string, eventKey string) {
		return senseEvent.Action,senseEvent.Key
	}
	
	senseConfig := SenseConfig{stopChan, *dir, 100 * time.Millisecond, keyGenFun}
	
	StartSensorPipeline("file", []string{"deduplicator"} , "console", senseConfig)
	
	<- stopChan
}
