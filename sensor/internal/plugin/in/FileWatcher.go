package in

import (
	. "github.com/gsharma85/go/sensor/internal/data"
	"log"
	"time"
	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	Watcher *fsnotify.Watcher
	OutChan chan SenseEvent
	StopSignal chan struct{}
}

func (fw FileWatcher) Watch() chan SenseEvent {
	go func() {
		for {
			select {
			case event, open := <-fw.Watcher.Events:
				if !open {
					log.Println("Event channel is closed by underlying directory watcher. Closing FileWatcher channel as well.")
				    close(fw.OutChan)
				    return
				}
				senseEvent := SenseEvent{event.Name, SUCCESS, event.Op.String(), time.Now()}
				fw.OutChan <- senseEvent
			case err, open := <- fw.Watcher.Errors:
				if !open {
					log.Println("Event channel is closed by underlying directory watcher. Closing FileWatcher channel as well.")
				    close(fw.OutChan)
				}
				log.Println("Error received from directory watcher: %s", err)
				senseEvent := SenseEvent{"Error", FAILURE, "Error", time.Now()}
				fw.OutChan <- senseEvent
			case _, open := <- fw.StopSignal:
				if !open {
				log.Println("Close channel requested. Closing FileWatcher.")
				    close(fw.OutChan)
				    fw.Watcher.Close()
				}
				
			}
		}
		
	}()
	
	return fw.OutChan
}

func CreateFileWatcher(senseConfig SenseConfig) (FileWatcher,bool) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal("Unable to create File Watcher. Error: %s", err)
		return FileWatcher{}, false
	}
	
	outChan := make(chan SenseEvent)
	fw := FileWatcher{watcher, outChan, senseConfig.StopSignal}
	
	if senseConfig.ActorSystemConfigData != nil {
		fw.Watcher.Add(senseConfig.ActorSystemConfigData.Address)
	} else {
		fw.Watcher.Add(senseConfig.DirName)
	}
	
	return fw, true
}