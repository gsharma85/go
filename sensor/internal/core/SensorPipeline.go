package core

import (
	. "github.com/gsharma85/go/sensor/internal/data"
	. "github.com/gsharma85/go/sensor/internal/plugin/in"
	. "github.com/gsharma85/go/sensor/internal/plugin/out"
	. "github.com/gsharma85/go/sensor/internal/plugin/transform"
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

func StartSensorPipeline(watcherName string, transformerNames []string, depositorName string, senseConfig SenseConfig) {
	watcher := getWatcher(watcherName, senseConfig)
	inChan := watcher.Watch()
	
	outChannel := inChan
	for _,transformer := range getTransformers(transformerNames, senseConfig) {
		outChannel = transformer.Transform(inChan)
	}
	depositor := getDepositor(depositorName, senseConfig)
	depositor.Deposit(outChannel)
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

func getTransformers(names [] string, senseConfig SenseConfig) []Transformer{
	transformers := make([]Transformer,0)
	for _,name := range names {
		transformers = append(transformers, getTransformer(name, senseConfig))
	}
	return transformers
}

func getTransformer(name string, senseConfig SenseConfig) Transformer{
	switch name {
	case "deduplicator": 
		transformer, ok := CreateDeDuplicator(senseConfig)
		if !ok {
			log.Fatal("Exiting application due to errors while creating deduplicator.")
		}
		return transformer
	}
	return nil
}


