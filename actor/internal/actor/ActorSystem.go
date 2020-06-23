package actor

import (
	"io/ioutil"
	"encoding/json"
	"github.com/gsharma85/go/dataflow/pkg/data"
	"github.com/gsharma85/go/actor/internal/utils"
)

type ActorSystem struct {
	Name string
	Address string
	Actors map[string]*Actor
}

func NewActorSystem(configFile string, logfile string, generateCommandHandlers func() map[string]func(Command, State) Response, generateTimeCommands func(actorConfig *data.ActorConfig) map[string]Command) chan Command {

	logger = utils.CreateFileLogger(logfile)
	stopAllActorsSignal := make(chan struct{})
	
	actorSystemConfig := parseActorConfig(configFile)
    
    actors := make(map[string]*Actor)
    
    commandHandlerMap := generateCommandHandlers()
    // create Actors
	for _, actorConfig := range actorSystemConfig.ActorConfigs {
	    actor := NewActor(actorConfig.Name, actorConfig.Address, commandHandlerMap, generateTimeCommands(actorConfig), stopAllActorsSignal, logger)
	    actors[actorConfig.Address] = actor
    }    
    
    // Create listen Actor response go routines
	
	// Create passivate Actor logic
	
	// Listen to commands from outer world
	
	// Create send command to Actors go routines
	return createActorSystemCommandChannel(actors)
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

func createActorSystemCommandChannel(actors map[string]*Actor) chan Command {
	commandChan := make(chan Command)
	for i:= 0; i < 10; i++ {
		go func() {
			for {
				command := <-commandChan
				actor := actors[command.ActorAddress]
				actor.InChan <- command
			}
		}()
	}
	return commandChan
}