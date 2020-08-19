package out

import (
	"log"
	"google.golang.org/grpc"
	"context"
	"time"
	"os"
	"fmt"
	"strings"
	grpcservice "github.com/gsharma85/go/dataflow/pkg/grpc"
	"github.com/gsharma85/go/dataflow/pkg/data"
	sensedata "github.com/gsharma85/go/sensor/internal/data"
	"github.com/gsharma85/go/sensor/internal/utils"
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
				if event.Action == "CREATE" {
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
				}			
			
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
	
	actorServiceHost,exists := os.LookupEnv("ACTOR_SERVICE_HOST")
	
	if !exists {
		actorHost,_ := os.Hostname()
		actorHostIP := utils.GetIpAddress(actorHost)
		log.Printf("ACTOR_SERVICE_HOST property not set. Using %s as default grpc port.", actorHostIP)
		actorServiceHost = actorHostIP;
	}
	
	actorServicePortStr,exists := os.LookupEnv("ACTOR_SERVICE_PORT")
	
	if !exists {
		log.Printf("ACTOR_SERVICE_PORT property not set. Using 3000 as default grpc port.")
		actorServicePortStr = "3000";
	}
	
	log.Printf("Dialing grpc connection to Actor System %s", fmt.Sprintf("%s:%s",actorServiceHost,actorServicePortStr))
	conn, dialErr := grpc.Dial(fmt.Sprintf("%s:%s",actorServiceHost,actorServicePortStr), grpc.WithInsecure())
	
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

