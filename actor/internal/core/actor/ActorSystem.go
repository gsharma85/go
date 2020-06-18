package actor

import (
	"github.com/gsharma85/go/actor/internal/data/config"
	"github.com/gsharma85/go/actor/internal/data/event"
	"github.com/gsharma85/go/actor/internal/utils"
	"log"
	"time"
)

type ActorSystem struct {
	Name string
	Address string
	Actors map[string]*Actor
}

var logger *log.Logger

func NewActorSystem(configFile string, logfile string) chan Command {

	logger = utils.CreateFileLogger(logfile)
	
    actorSystemConfig := parse(configFile)
    actors := make(map[string]*Actor)
    
    for _, actorConfig := range actorSystemConfig.ActorConfigs {
//	    actor := NewActor()
    }
    
    
	//actorMap := make(map[string]*Actor)
	
	// create Actors
	
	// Create send command to Actors go routines
	
	// Create listen Actor response go routines
	
	// Create passivate Actor logic
	
	// Listen to commands from outer world
	
	return nil
	
}

func parse(filePath string) config.ActorSystemConfig {
	return config.ActorSystemConfig{}
}

func getCommandHandlerMap() map[string]func(Command, State) Response {
	commandMap := make(map[string]func(Command, State) Response)
	commandMap["HandleCreateFileEvent"] = handleCreateUpdateFileCommand
	commandMap["HandleUpdateFileEvent"] = handleCreateUpdateFileCommand
	commandMap["HandleCheckFileArrival"] = handleCheckFileArrivalCommand
	
	return commandMap
}

func getTimedCommands(actorConfig config.ActorConfig) map[string]Command {
	timedcommandMap := make(map[string]Command)
	commandTime := time.Now()
	checkFileArrivalCommand := Command{"HandleCheckFileArrival", commandTime, nil}
	timedcommandMap["HandleCheckFileArrival"] = checkFileArrivalCommand
	return timedcommandMap
}


func handleCreateUpdateFileCommand(command Command, state State) Response {
	fileEvent := command.Payload.(event.FileEvent)
	state.Data["MostRecentEvent"] = fileEvent
	count, ok := state.Data["FileCreateOrUpdateCount"]
	
	if !ok {
		count = 0
	}
	count = count + 1;
	
	alerts, ok := state.Data["Alerts"]
	
	if !ok {
		alerts = make[[]string]
	}
	
	if count > 1 {
		alerts = append(alerts, "File create/updated more than once")
		logger.Println("Alert on file %s - %s", fileEvent.Name, "File create/updated more than once")
	}
	
	state.Data["Alerts"] = alerts
	
	response := Response{command.Name, "Processed"}
	return response
}

func handleCheckFileArrivalCommand(command Command, state State) Response {
	fileEvent := command.Payload.(event.FileEvent)
	state.Data["MostRecentEvent"] = fileEvent
	
	alerts, ok := state.Data["Alerts"]
	
	if !ok {
		alerts = make[[]string]
	}
	
	count, ok := state.Data["FileCreateOrUpdateCount"]
	
	if !ok || count == 0 {
		alerts = append(alerts, "File not received by configured time.")
		logger.Println("Alert on file %s - %s", fileEvent.Name, "File not received by configured time.")		
	}
	
	state.Data["Alerts"] = alerts
	
	response := Response{command.Name, "Processed"}
	return response
}