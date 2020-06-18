package actor

import (
	"time"
	"log"
)

type State struct {
	Data map[string]interface{}
}


type Processor struct {
	CommandProcessor map[string]func(Command, State) Response
	TimedCommands map[string]Command
}

type Command struct {
	Name string
	CommandTime time.Time
	Payload interface{}
}

type Response struct {
	CommandName string
	Response interface{}
}

type Actor struct {
	Name string
	Address string
	processor Processor
	InChan chan Command
	OutChan chan Response
	State State
	stopActorSystemSignal chan struct{} 
}

func NewActor(name string, address string, commandProcessors map[string]func(Command, State) Response, timedCommands map[string]Command, stopActorSystemSignal chan struct{}) *Actor {
	inChan := make(chan Command)
	outChan := make(chan Response)
	state := State{make(map[string]interface{})}
	processor := Processor{commandProcessors, timedCommands}
	actor := Actor{name, address, processor, inChan, outChan, state, stopActorSystemSignal}
	
	// Load generic commands like buildState, passivateActor, getState, pushStateOnChannel
	
	// Start Actor command processor
	closeActorSelfGoRoutines := make(chan struct{})
	actor.createExternalCommandRoutine(closeActorSelfGoRoutines)
	
	
	// Start actor self commands
	actor.createSelfCommandRoutines(closeActorSelfGoRoutines)
	
	return &actor
}

func (actorRef *Actor) createExternalCommandRoutine(closeActorSelfGoRoutines chan struct{}) {
	go func(actor *Actor) {
		for {
			select {
				case Command, open := <-actor.InChan:
				if !open {
					close(actor.InChan)
					close(actor.OutChan)
					return
				}
				
				processor := actor.processor.CommandProcessor[Command.Name]
				response := processor(Command, actor.State)
				actor.OutChan <- response
				
				// Close all the self go routines if this was a passivate command
				if Command.Name == "passivateActor" {
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

func (actorRef *Actor) createSelfCommandRoutines(closeActorSelfGoRoutines chan struct{}) {
	if actorRef.processor.TimedCommands != nil && len(actorRef.processor.TimedCommands) > 0 {
		for name, command := range actorRef.processor.TimedCommands {
			go func(name string, command Command, actorInChan chan Command, closeActorSelfGoRoutines chan struct{}) {
				
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
			}(name, command, actorRef.InChan, closeActorSelfGoRoutines)
		}
	}
}
