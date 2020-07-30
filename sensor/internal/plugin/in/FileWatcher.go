package in

import (
	. "github.com/gsharma85/go/sensor/internal/data"
	"log"
	"time"
	"strings"
	"github.com/fsnotify/fsnotify"
)

type FileWatcher struct {
	Watcher *fsnotify.Watcher
	OutChan chan SenseEvent
	StopSignal chan struct{}
}

type filePath struct {
	Address string
	Childs []string
}

var watchableDirs = make(map[string]string)

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
				path := strings.ReplaceAll(event.Name, "\\", "/")
				_ , ok := watchableDirs[path]
				if ok {
					fw.Watcher.Add(path)
				}
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
		
		filePaths := make(map[string]*filePath)
		for _, actorConfig := range senseConfig.ActorSystemConfigData.ActorConfigs {
		    file := filePath{actorConfig.Address, make([]string,0)}
		    filePaths[actorConfig.Address] = &file
	    }    
	
		for address, _ := range filePaths {
		    lastIndex := strings.LastIndex(address, "/")
		    chars := strings.Split(address, "")
	    
		    log.Printf("Last index: %d, Length: %d", lastIndex, len(address))
		    parentPath := strings.Join(chars[:lastIndex], "")
	    
		    if address != senseConfig.ActorSystemConfigData.Address {
			    log.Printf("Adding %s as child to %s", address, parentPath)
			    filePaths[parentPath].Childs = append(filePaths[parentPath].Childs, address)
		    }	    
	    }
		
		for address, file := range filePaths {
			if len(file.Childs) > 0 {
				watchableDirs[address] = address
			}
		}
	} else {
		fw.Watcher.Add(senseConfig.DirName)
	}
	
	return fw, true
}