package transform

import (
	"time"
	"log"
	. "github.com/gsharma85/go/sensor/internal/data"
)

type DeDuplicator struct {
	DeDupInterval time.Duration
	OutChannel chan SenseEvent
	StopSignal chan struct{}
	DeDupKeyGen func(senseEvent SenseEvent) (groupByKey string, eventKey string)
}

func (deDup DeDuplicator) Transform(inChan chan SenseEvent) chan SenseEvent {
	go func() {
		eventMap := make(map[string]map[string][]SenseEvent)
		ticker := time.NewTicker(deDup.DeDupInterval)
		for {
			select {
			case sensedEvent, open:= <- inChan:
			if !open {
				close(deDup.OutChannel)
				return
			}
			log.Println("Got an event for deduplication: %s", sensedEvent)
			groupByKey,eventKey := deDup.DeDupKeyGen(sensedEvent)
			eventMapForKey,exists := eventMap[eventKey]
			if !exists {
				eventMapForKey = make(map[string][]SenseEvent)
				eventMap[eventKey] = eventMapForKey
			}
			
			eventsForGroupKey,exists := eventMapForKey[groupByKey]
			if !exists {
				eventsForGroupKey = make([] SenseEvent,0)
			}
			
			eventsForGroupKey = append(eventsForGroupKey, sensedEvent)
			eventMapForKey[groupByKey] = eventsForGroupKey	
				
			case _, open := <- ticker.C:
			if !open {
				close(deDup.OutChannel)
				return
			}
			if len(eventMap) <= 0 {
				break
			}
			log.Println("Start deduplication process.")
			
			deDuplicatedEvents := make([] SenseEvent,0)
			for eventKey, eventMapForKey := range eventMap {
				log.Println("Deduplicating events for key: %s",eventKey)
				for _,eventsForGroupKey := range eventMapForKey {
					deDuplicatedEvents = append(deDuplicatedEvents, eventsForGroupKey[0])
				}
				delete(eventMap,eventKey)
			}
			go func() {
				for _, event := range deDuplicatedEvents {
					deDup.OutChannel <- event
				}
			}()
			
			case _, open := <- deDup.StopSignal:
			if !open {
				close(deDup.OutChannel)
				return
			}
			}
		}
	}()
	
	return deDup.OutChannel
}

func CreateDeDuplicator(senseConfig SenseConfig) (DeDuplicator,bool) {
	outChannel := make(chan SenseEvent)
	deDuplicator := DeDuplicator{senseConfig.DeDupInterval, outChannel, senseConfig.StopSignal, senseConfig.DeDupKeyGen}
	return deDuplicator,true
}