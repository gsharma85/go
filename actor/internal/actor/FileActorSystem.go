package actor

import (
	"github.com/gsharma85/go/dataflow/pkg/data"
	"github.com/gsharma85/go/actor/internal/eventlistener/grpc"
	"log"
	"time"
)

var actors map[string]*Actor

func NewFileActorSystem(configFile string, logfile string) {
	commandInChan := NewActorSystem(configFile, logfile, getCommandHandlerMap, getTimedCommands)
	startFileEventReceptor(commandInChan)
}

func getCommandHandlerMap() map[string]func(Command, State) Response {
	commandMap := make(map[string]func(Command, State) Response)
	commandMap["HandleCreateOrUpdateFileEvent"] = handleCreateUpdateFileCommand
	commandMap["HandleCheckFileArrival"] = handleCheckFileArrivalCommand
	
	return commandMap
}

func getTimedCommands(actorConfig *data.ActorConfig) map[string]Command {
	timedcommandMap := make(map[string]Command)
	commandTime, err := time.Parse(time.Kitchen, actorConfig.CompleteBy)
	
	if err != nil {
		logger.Printf("Error while parsing complete by time from config: %s", err)
		log.Fatal("Error while parsing complete by time from config: %s", err)
	}
	
	checkFileArrivalCommand := Command{"HandleCheckFileArrival", actorConfig.Address, commandTime, nil}
	timedcommandMap["HandleCheckFileArrival"] = checkFileArrivalCommand
	return timedcommandMap
}


func handleCreateUpdateFileCommand(command Command, state State) Response {
	fileEvent := command.Payload.(*data.FileEvent)
	state.Data["MostRecentEvent"] = fileEvent
	value, ok := state.Data["FileCreateOrUpdateCount"]
	
	var count int32
	if !ok {
		count = 0
	} else {
		count = value.(int32)
	}
	
	count = count + 1;
	
	value, ok = state.Data["Alerts"]
	
	var alerts []string
	if !ok {
		alerts = make([]string,0)
	} else {
		alerts = value.([]string)
	}
	
	if count > 1 {
		alerts = append(alerts, "File create/updated more than once")
		logger.Printf("Alert on file %s - %s", fileEvent.Name, "File create/updated more than once")
	}
	
	state.Data["Alerts"] = alerts
	
	response := Response{command.Name, "Processed"}
	return response
}

func handleCheckFileArrivalCommand(command Command, state State) Response {
	
	log.Printf("Start processing file arrival check command.")
	
	value, ok := state.Data["Alerts"]
	
	var alerts []string
	
	if !ok {
		alerts = make([]string,0)
	} else {
		alerts = value.([]string)
	}
	
	count, ok := state.Data["FileCreateOrUpdateCount"]
	
	if !ok || count == 0 {
		alerts = append(alerts, "File not received by configured time.")
		log.Printf("Alert on file %s - %s", command.ActorPath, "File not received by configured time.")
		logger.Printf("Alert on file %s - %s", command.ActorPath, "File not received by configured time.")		
	}
	
	state.Data["Alerts"] = alerts
	
	response := Response{command.Name, "Processed"}
	return response
}

func startFileEventReceptor(commandInChan chan Command) {
	fileEventInChan := grpc.StartListener()
	
	go func() {
		for {
			fileEvent := <- fileEventInChan
			log.Printf("Got file event: %s", fileEvent)
			command := Command{fileEvent.Name, fileEvent.ActorPath, time.Now(), fileEvent}
			commandInChan <- command
		}
	}()
	
}