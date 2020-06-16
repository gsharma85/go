package actor

import (
	"time"
	"log"
)

type ActorState struct {
	data map[string]interface{}
}


type ActorProcessor struct {
	CommandProcessor map[string]func(ActorCommand, ActorState) ActorResponse
	TimedCommands map[string]ActorCommand
}

type ActorCommand struct {
	Name string
	CommandTime time.Time
	Payload interface{}
}

type ActorResponse struct {
	CommandName string
	Response interface{}
}

type Actor struct {
	Name string
	Address string
	processor ActorProcessor
	InChan chan ActorCommand
	OutChan chan ActorResponse
	State ActorState
	stopActorSystemSignal chan struct{} 
}

func CreateActor(name string, address string, commandProcessors map[string]func(ActorCommand, ActorState) ActorResponse, TimedCommands map[string]ActorCommand, stopActorSystemSignal chan struct{}) *Actor {
	inChan := make(chan ActorCommand)
	outChan := make(chan ActorResponse)
	state := ActorState{make(map[string]interface{})}
	processor := ActorProcessor{commandProcessors, TimedCommands}
	actor := Actor{name, address, processor, inChan, outChan, state, stopActorSystemSignal}
	
	// Load generic commands like buildActorState, passivateActor, getState
	
	// Start Actor command processor
	closeActorSelfGoRoutines := make(chan struct{})
	go func(actor *Actor) {
		for {
			select {
				case actorCommand, open := <-actor.InChan:
				if !open {
					close(inChan)
					close(outChan)
					return
				}
				
				processor := actor.processor.CommandProcessor[actorCommand.Name]
				response := processor(actorCommand, actor.State)
				outChan <- response
				
				// Close all the self go routines if this was a passivate command
				if actorCommand.Name == "passivateActor" {
					close(closeActorSelfGoRoutines)
					return
				}
				
				case _, open := <- actor.stopActorSystemSignal:
				if !open {
					close(inChan)
					close(outChan)
					return
				}
			}
		}
	}(&actor)
	
	
	// Start actor self commands
	if TimedCommands != nil && len(TimedCommands) > 0 {
		for name, command := range TimedCommands {
			ticker := time.NewTicker(500 * time.Millisecond)
			go func(name string, command ActorCommand, actorInChan chan ActorCommand, closeActorSelfGoRoutines chan struct{}) {
				for {
					select {
						case time,_ := <- ticker.C:
						log.Println("Self command executed at %s", time)
						actorInChan <- command
						
						case _, open := <- closeActorSelfGoRoutines:
						if !open {
							return
						}
					}
				}
			}(name, command, actor.InChan, closeActorSelfGoRoutines)
		}
	}
	
	return &actor
}
