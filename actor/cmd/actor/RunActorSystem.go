package main

import (
	"flag"
	"github.com/gsharma85/go/actor/internal/actor"
	"github.com/gsharma85/go/actor/internal/receptor"
)

func main() {
	
	configFile := flag.String("cf", "config file", "actor config file")
	logfile := flag.String("lf", "log file", "where event logs will be writtern")
	flag.Parse()
	
	stopChan := make(chan struct{})
	
	commandInChan := actor.NewFileActorSystem(*configFile, *logfile)
	receptor.StartFileEventReceptor(commandInChan)
	
	<- stopChan
}
