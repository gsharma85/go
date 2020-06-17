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

func CreateActor(name string, address string, commandProcessors map[string]func(ActorCommand, ActorState) ActorResponse, timedCommands map[string]ActorCommand, stopActorSystemSignal chan struct{}) *Actor {
	inChan := make(chan ActorCommand)
	outChan := make(chan ActorResponse)
	state := ActorState{make(map[string]interface{})}
	processor := ActorProcessor{commandProcessors, timedCommands}
	actor := Actor{name, address, processor, inChan, outChan, state, stopActorSystemSignal}
	
	// Load generic commands like buildActorState, passivateActor, getState, pushStateOnChannel
	
	// Start Actor command processor
	closeActorSelfGoRoutines := make(chan struct{})
	createExternalCommandRoutine(&actor, closeActorSelfGoRoutines)
	
	
	// Start actor self commands
	createSelfCommandRoutines(timedCommands, actor.InChan, closeActorSelfGoRoutines)
	
	return &actor
}

func createExternalCommandRoutine(actorRef *Actor, closeActorSelfGoRoutines chan struct{}) {
	go func(actor *Actor) {
		for {
			select {
				case actorCommand, open := <-actor.InChan:
				if !open {
					close(actor.InChan)
					close(actor.OutChan)
					return
				}
				
				processor := actor.processor.CommandProcessor[actorCommand.Name]
				response := processor(actorCommand, actor.State)
				actor.OutChan <- response
				
				// Close all the self go routines if this was a passivate command
				if actorCommand.Name == "passivateActor" {
					close(closeActorSelfGoRoutines)
					return
				}
				
				case _, open := <- actor.stopActorSystemSignal:
				if !open {
					close(actor.InChan)
					close(actor.OutChan)
					return
				}
			}
		}
	}(actorRef)
}

func createSelfCommandRoutines(timedCommands map[string]ActorCommand, actorInChan chan ActorCommand, closeActorSelfGoRoutines chan struct{}) {
	if timedCommands != nil && len(timedCommands) > 0 {
		for name, command := range timedCommands {
			go func(name string, command ActorCommand, actorInChan chan ActorCommand, closeActorSelfGoRoutines chan struct{}) {
				
				timerResetFunc := func(scheduledOn time.Time) *time.Timer{
					triggerDelay := scheduledOn.Sub(time.Now())
					
					var nextTriggerOn time.Duration
					if triggerDelay <= 0 {
						nextTriggerOn = time.Until(scheduledOn.Add(time.Hour * 24))
					} else {
						nextTriggerOn = time.Until(scheduledOn)
					}
					
					log.Println("Next timer for command: %s will be in %s", name, nextTriggerOn)
					
					return time.NewTimer(nextTriggerOn)
				}
			
				timer := timerResetFunc(command.CommandTime)
			
				for {
					select {
						case tickTime,_ := <- timer.C:
						log.Println("Self command executed at %s", tickTime)
						actorInChan <- command
						timer = timerResetFunc(time.Now().Add(time.Hour * 24))
						
						case _, open := <- closeActorSelfGoRoutines:
						if !open {
							return
						}
					}
				}
			}(name, command, actorInChan, closeActorSelfGoRoutines)
		}
	}
}
