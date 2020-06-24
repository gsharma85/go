package main

import (
	. "github.com/gsharma85/go/sensor/internal/core"
	. "github.com/gsharma85/go/sensor/internal/data"
	"time"
	"flag"
	"encoding/json"
	"io/ioutil"
	"log"
	dataflow "github.com/gsharma85/go/dataflow/pkg/data"
)

func main() {
	
	dir := flag.String("d", "directory", "directory to watch")
	configFile := flag.String("cf", "config file", "actor config file")
	logfile := flag.String("lf", "log file", "where event logs will be writtern")
	flag.Parse()
	
	var actorSystemConfig *dataflow.ActorSystemConfig
	if configFile != nil {
		actorSystemConfig = parseActorConfig(*configFile)
	}
	
	stopChan := make(chan struct{})
	keyGenFun := func(senseEvent SenseEvent) (groupByKey string, eventKey string) {
		return senseEvent.Action,senseEvent.Key
	}
	
	senseConfig := SenseConfig{stopChan, *dir, 100 * time.Millisecond, keyGenFun, *logfile, actorSystemConfig}
	
	StartSensorPipeline("file", []string{"deduplicator"} , "grpcactorsystem", senseConfig)
	
	<- stopChan
}

func parseActorConfig(filePath string) *dataflow.ActorSystemConfig {
	file, err := ioutil.ReadFile(filePath)
	
	if err != nil {
		log.Println("Error reading file :%s. Error is: %s", filePath, err)
		return nil
	} 
	
	var actorSystemConfig dataflow.ActorSystemConfig 
	err = json.Unmarshal([]byte(file), &actorSystemConfig)
	
	if err != nil {
		log.Println("Error marshalling into ActorSystemConfig. Error is: %s", err)
		return nil
	} 
	return &actorSystemConfig
}