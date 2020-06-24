package out

import (
	"log"
	"google.golang.org/grpc"
	"context"
	"time"
	"strings"
	grpcservice "github.com/gsharma85/go/dataflow/pkg/grpc"
	"github.com/gsharma85/go/dataflow/pkg/data"
	sensedata "github.com/gsharma85/go/sensor/internal/data"
)

type GrpcOutToActorSystem struct {
	ActorSystemPath string
	StopSignal chan struct{}
	GrpcOutStream grpcservice.FileMonitoringActorSystemService_HandleFileEventClient
}

func (grpcOut GrpcOutToActorSystem) Deposit(inChan chan sensedata.SenseEvent) {
	go func() {
		closeRoutine := false
		for {
			select {
				case event, _ := <- inChan:
				log.Println("Got file sense event : %s", event)
				timeStr := event.Time.Format(time.RFC1123)
				actorPath := strings.ReplaceAll(event.Key, "\\", "/")
				fileEvent :=  data.FileEvent{}
				fileEvent.Name  = "HandleCreateOrUpdateFileEvent"
				fileEvent.ActorSystemPath = grpcOut.ActorSystemPath
				fileEvent.ActorPath = actorPath
				fileEvent.Action = event.Action
				fileEvent.Time = timeStr
				log.Println("Sending event to Actor system: %s", fileEvent)
				grpcOut.GrpcOutStream.Send(&fileEvent)
			
				case _, open := <- grpcOut.StopSignal:
				if !open {
					closeRoutine = true
				}
			}
			if closeRoutine {
				return
			}
		}
	} ()
}

func CreateGrpcToActorSystemDepositor(senseConfig sensedata.SenseConfig) (GrpcOutToActorSystem,bool) {
	log.Printf("Dialing grpc connection to Actor System.")
	
	conn, dialErr := grpc.Dial("127.0.0.1:7773", grpc.WithInsecure())
	
	if dialErr != nil {
		log.Fatal("Unable to make grpc connection to Actor System: %s", dialErr)
		return GrpcOutToActorSystem{}, false
	}
	
	client := grpcservice.NewFileMonitoringActorSystemServiceClient(conn)
	clientStream, err := client.HandleFileEvent(context.Background())
	
	if err != nil {
		log.Fatal("Error opening client stream to Actor System: %s", err)
	}
	
	return GrpcOutToActorSystem{senseConfig.ActorSystemConfigData.Address, senseConfig.StopSignal, clientStream}, true
}

