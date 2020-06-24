package receptor

import (
	"github.com/gsharma85/go/actor/internal/actor"
	"github.com/gsharma85/go/actor/internal/eventlistener/grpc"
	"time"
	"log"
)

func StartFileEventReceptor(commandInChan chan actor.Command) {
	fileEventInChan := grpc.StartListener()
	
	go func() {
		for {
			fileEvent := <- fileEventInChan
			log.Printf("Got file event: %s", fileEvent)
			command := actor.Command{fileEvent.Name, fileEvent.ActorPath, time.Now(), fileEvent}
			commandInChan <- command
		}
	}()
	
}
