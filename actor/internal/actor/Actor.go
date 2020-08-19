package actor

import (
	"time"
	"log"
	"hash/fnv"
	"fmt"
	"io/ioutil"
	"encoding/json"
	"os"
)

type State struct {
	Alerts []string
	Status map[string]bool
	ChildStatus map[string]bool
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
	Type string
	Address string
	Childs []string
	State State
	processor Processor
	InChan chan Command
	SystemChan chan Command
	stopActorSystemSignal chan struct{} 
	logger *log.Logger
	DbDir string
}

type ActorBuilder struct {
	Name string
	Type string
	Address string
	Childs []string
	TimedCommands map[string]Command
	DbDir string
}

func NewActorBuilder(name string, address string, timedCommands map[string]Command, DbDir string) ActorBuilder {
	actorBuilder := ActorBuilder{}
	actorBuilder.Name = name
	actorBuilder.Address = address
	actorBuilder.Childs = make([]string, 0)
	actorBuilder.TimedCommands = timedCommands
	actorBuilder.DbDir = DbDir
	return actorBuilder
}

func(ab *ActorBuilder) Build(commandProcessors map[string]func(Command, State) Response, systemChan chan Command, stopActorSystemSignal chan struct{}, logger *log.Logger) *Actor {
	inChan := make(chan Command)
	state := State{make([]string,0),make(map[string]bool), nil}
	processor := Processor{commandProcessors, ab.TimedCommands}
	actor := Actor{ab.Name, ab.Type, ab.Address, ab.Childs, state, processor, inChan, systemChan, stopActorSystemSignal, logger, ab.DbDir}
	
	if len(actor.Childs) != 0 {
		childStatus := make(map[string]bool)
		for _, cAddress := range actor.Childs {
			childStatus[cAddress] = false
		}
		actor.State.ChildStatus = childStatus
	}
	
	actor.State.Status["nodeComplete"] = false
	actor.State.Status["status"] = false
	actor.loadState()
	// Load generic commands like buildState, passivateActor, getState, pushStateOnChannel
	processor.CommandProcessor["ChildCompleteEvent"] = childStateCompleteProcessor
	
	// Start Actor command processor
	closeActorSelfGoRoutines := make(chan struct{})
	actor.createExternalCommandRoutine(closeActorSelfGoRoutines)
	
	
	// Start actor self commands
	actor.createSelfCommandRoutines(closeActorSelfGoRoutines)
	
	return &actor
}

func (actor *Actor) loadState() {
	stateFileName := actor.getActorStateFileName()
	log.Printf("Check if state file exists")
	if stateFileExists(stateFileName) {
		log.Printf("Loading state file.")
		file, err := ioutil.ReadFile(stateFileName)
	
		if err != nil {
			logger.Println("Error reading file :%s. Error is: %s", stateFileName, err)
		} 
	
		log.Printf("Unmarshalling state file.")
		var state State 
		err = json.Unmarshal([]byte(file), &state)
		actor.State = state
		if err != nil {
			logger.Println("Error Unmarshalling into Actor.State. Error is: %s", err)
		}
	}
}

func (actor *Actor) persistState() {
	stateFileName := actor.getActorStateFileName()
	byteArr, err := json.Marshal(actor.State)
	if err != nil {
		logger.Println("Error marshalling actor %s state. %s", actor.Address, err)
	}
	err = ioutil.WriteFile(stateFileName, byteArr, 0600)
	if err != nil {
		logger.Println("Error writing actor %s state. %s", actor.Address, err)
	}
}

func (actor *Actor) getActorStateFileName() string {
	h := fnv.New32a()
    h.Write([]byte(actor.Address))
    return fmt.Sprintf("%s/actor-%d.db", actor.DbDir, h.Sum32())
}

func stateFileExists(name string) bool {
    if _, err := os.Stat(name); err != nil {
        if os.IsNotExist(err) {
            return false
        }
    }
    return true
}

func childStateCompleteProcessor(cmd Command, state State) Response {
	childAddress := cmd.Payload.(string)
	state.ChildStatus[childAddress] = true
	return Response{cmd.Name, "Processed"}
}

func (actorRef *Actor) createExternalCommandRoutine(closeActorSelfGoRoutines chan struct{}) {
	go func(actor *Actor) {
		for {
			select {
				case cmd, open := <-actor.InChan:
				if !open {
					close(actor.InChan)
					return
				}
				
				log.Printf("Got command %s for actor %s", cmd.Name, actor.Address)
				logger.Printf("Got command %s for actor %s", cmd.Name, actor.Address)
				
				processor := actor.processor.CommandProcessor[cmd.Name]
				response := processor(cmd, actor.State)
				log.Printf("Response after running command: %s", response)
				
				
				isComplete := actor.stateCheck()
				
				actor.State.Status["status"] = isComplete
				
				actor.persistState()
				if isComplete {
					log.Println("Sending child complete command for: %s", actor.Address)
					actor.SystemChan <- Command{"ChildCompleteEvent", "parent", time.Now(), actor.Address}
				}
				
				// Close all the self go routines if this was a passivate command
				if cmd.Name == "passivateActor" {
					close(closeActorSelfGoRoutines)
					return
				}
				
				case _, open := <- actor.stopActorSystemSignal:
				if !open {
					close(actor.InChan)
					return
				}
			}
		}
	}(actorRef)
}

func (actorRef *Actor) stateCheck() bool {
	if len(actorRef.Childs) == 0 {
		return actorRef.State.Status["nodeComplete"]
	} else {
		childComplete := actorRef.State.ChildStatus
		for _, isComplete := range childComplete {
			if !isComplete {
				return false
			}
		}
		return actorRef.State.Status["nodeComplete"]
	}
	
	return false
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
