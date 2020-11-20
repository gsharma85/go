package actor

import (
	"io/ioutil"
	"encoding/json"
	"github.com/gsharma85/go/dataflow/pkg/data"
	"github.com/gsharma85/go/actor/internal/utils"
	"log"
	"strings"
)

var logger *log.Logger

type ActorSystem struct {
	Name string
	RootActorAddress string
	Actors map[string]*Actor
	ExternalCommandChan chan Command
	InternalCommndChan chan Command
}

func NewActorSystem(configFile string, logfile string, dbDir string, generateCommandHandlers func() map[string]func(Command, *State) Response, generateTimeCommands func(actorConfig *data.ActorConfig) map[string]Command) ActorSystem {

	logger = utils.Logger(logfile)
	stopAllActorsSignal := make(chan struct{})
	
	actorSystemConfig := parseActorConfig(configFile)
    
    systemChan := make(chan Command)
    actors := make(map[string]*Actor)
    actorBuilders := make(map[string]*ActorBuilder)
    
    // create Actors builders
	for _, actorConfig := range actorSystemConfig.ActorConfigs {
	    actorBuilder := NewActorBuilder(actorConfig.Name, actorConfig.Address, generateTimeCommands(actorConfig), dbDir)
	    actorBuilders[actorConfig.Address] = &actorBuilder
    }    
	
	for address, _ := range actorBuilders {
	    lastIndex := strings.LastIndex(address, "/")
	    chars := strings.Split(address, "")
	    
	    log.Printf("Last index: %d, Length: %d", lastIndex, len(address))
	    parentPath := strings.Join(chars[:lastIndex], "")
	    
	    if address != actorSystemConfig.Address {
		    log.Printf("Adding %s as child to %s", address, parentPath)
		    actorBuilders[parentPath].Childs = append(actorBuilders[parentPath].Childs, address)
	    }    
    }
    
    // Create Actors
    for _ , actorBuilder := range actorBuilders {
    	actor := actorBuilder.Build(generateCommandHandlers(), systemChan, stopAllActorsSignal, logger)
		actors[actor.Address] = actor
    }
    
    // Set root actor complete to true, events will come only on childs
    actors[actorSystemConfig.Address].State.Status["complete"] = true
    
    // Create passivate Actor logic
	
	// Listen to commands from outer world
	
	// Create send command to Actors go routines
	createActorSystemIntCommandChannel(actors, actorSystemConfig.Address, systemChan)
	return ActorSystem{actorSystemConfig.Name, actorSystemConfig.Address, actors, createActorSystemExtCommandChannel(actors), systemChan}
}

func parseActorConfig(filePath string) *data.ActorSystemConfig {
	file, err := ioutil.ReadFile(filePath)
	
	if err != nil {
		logger.Println("Error reading file :%s. Error is: %s", filePath, err)
		return nil
	} 
	
	var actorSystemConfig data.ActorSystemConfig 
	err = json.Unmarshal([]byte(file), &actorSystemConfig)
	
	if err != nil {
		logger.Println("Error marshalling into ActorSystemConfig. Error is: %s", err)
		return nil
	} 
	return &actorSystemConfig
}

func createActorSystemExtCommandChannel(actors map[string]*Actor) chan Command {
	commandChan := make(chan Command)
	for i:= 0; i < 10; i++ {
		go func() {
			for {
				command := <-commandChan
				actor := actors[command.ActorPath]
				
				if actor != nil {
					actor.InChan <- command
				}
			}
		}()
	}
	return commandChan
}

func createActorSystemIntCommandChannel(actors map[string]*Actor, actorSystemAddress string, systemChan chan Command) {
	for i:= 0; i < 10; i++ {
		go func() {
			for {
				command := <-systemChan
				
				if command.Name == "ChildCompleteEvent" && command.Payload.(string) != actorSystemAddress {
					parent := command.ActorPath
					log.Println("Got child complete command: %s", command)    
					if parent == "parent" {
						childAddress := command.Payload.(string)
						lastIndex := strings.LastIndex(childAddress, "/")
					    chars := strings.Split(childAddress, "")
					    parent = strings.Join(chars[:lastIndex], "")
					}
					
					actor := actors[parent]
					command.ActorPath = parent
					if actor != nil {
						log.Println("Forwarding child complete command: %s", command)
						actor.InChan <- command
					}
				}
				
			}
		}()
	}
}