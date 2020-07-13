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
	ActorPath string
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
	logger *log.Logger
}

func NewActor(name string, address string, commandProcessors map[string]func(Command, State) Response, timedCommands map[string]Command, stopActorSystemSignal chan struct{}, logger *log.Logger) *Actor {
	inChan := make(chan Command)
	outChan := make(chan Response)
	state := State{make(map[string]interface{})}
	processor := Processor{commandProcessors, timedCommands}
	actor := Actor{name, address, processor, inChan, outChan, state, stopActorSystemSignal, logger}
	
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
				
				log.Printf("Got command %s for actor %s", Command.Name, actorRef.Address)
				logger.Printf("Got command %s for actor %s", Command.Name, actorRef.Address)
				
				processor := actor.processor.CommandProcessor[Command.Name]
				response := processor(Command, actor.State)
				log.Printf("Response after running command: %s", response)
				
//				actor.OutChan <- response
				
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
				
				currentTime := time.Now()
				scheduledOn := time.Date(currentTime.Year(), currentTime.Month(), currentTime.Day(), command.CommandTime.Hour(), command.CommandTime.Minute(), command.CommandTime.Second(), command.CommandTime.Nanosecond(), currentTime.Location())
				timerResetFunc := func(scheduledOn time.Time) *time.Timer{
					triggerDelay := scheduledOn.Sub(time.Now())
					
					var nextTriggerOn time.Duration
					if triggerDelay <= 0 {
						nextTriggerOn = time.Minute * 1
					} else {
						nextTriggerOn = time.Until(scheduledOn)
					}
					
					log.Printf("Next timer for command: %s for actor %s will be in %s", name, actorRef.Address, nextTriggerOn)
					logger.Printf("Next timer for command: %s for actor %s will be in %s", name, actorRef.Address, nextTriggerOn)
					
					return time.NewTimer(nextTriggerOn)
				}
			
				timer := timerResetFunc(scheduledOn)
				// Run if time already passed
				triggerDelay := scheduledOn.Sub(time.Now())
				if triggerDelay <= 0 {
					log.Printf("Set up time in past. Running self command %s for actor %s.", name, actorRef.Address)
					logger.Printf("Set up time in past. Running self command %s for actor %s.", name, actorRef.Address)
					actorInChan <- command
				}
				
				for {
					select {
						case tickTime,_ := <- timer.C:
						log.Printf("Running self command %s for actor %s at time %s.", name, actorRef.Address, tickTime)
						logger.Printf("Running self command %s for actor %s at time %s", name, actorRef.Address, tickTime)
						actorInChan <- command
						timer = timerResetFunc(time.Now().Add(time.Minute * 1))
						
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
