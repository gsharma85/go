package core

import (
	. "github.com/gsharma85/go/sensor/internal/data"
	. "github.com/gsharma85/go/sensor/internal/plugin/in"
	. "github.com/gsharma85/go/sensor/internal/plugin/out"
	"log"
)

type Watcher interface {
	Watch() chan SenseEvent
}

type Transformer interface {
	Transform(inChan chan SenseEvent) chan SenseEvent
}

type Depositor interface {
	Deposit(inChan chan SenseEvent)
}

type SensorPipeline struct {
	watch Watcher
	transform []Transformer
	deposit Depositor
}

func StartSensorPipeline(watcherName string, depositorName string, senseConfig SenseConfig) {
	watcher := getWatcher(watcherName, senseConfig)
	inChan := watcher.Watch()
	
//	outChannel := inChannel
//	for _,transformer := range transformers {
//		outChannel = transformer.Transform(outChannel)
//	}
	depositor := getDepositor(depositorName, senseConfig)
	depositor.Deposit(inChan)
}

func getWatcher(name string, senseConfig SenseConfig) Watcher {
	switch name {
	case "file": 
		fw, ok := CreateFileWatcher(senseConfig)
		if !ok {
			log.Fatal("Exiting application due to errors while creating FileWatcher")
		}
		return fw
	}
	return nil
}

func getTransformers(names [] string, senseConfig SenseConfig) []Transformer{
	return nil
}

func getDepositor(name string, senseConfig SenseConfig) Depositor {
	switch name {
	case "console": 
		logger, ok := CreateConsoleDepositor(senseConfig)
		if !ok {
			log.Fatal("Exiting application due to errors while creating Console logger.")
		}
		return logger
	}
	return nil
}



