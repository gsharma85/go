package grpc

import (
	"github.com/gsharma85/go/dataflow/pkg/data"
	grpcservice "github.com/gsharma85/go/dataflow/pkg/grpc"
	"github.com/gsharma85/go/actor/internal/utils"
	"google.golang.org/grpc"
	"log"
	"net"
	"os"
	"fmt"
)

type fileEventInEndPoint struct {
	EventInChan chan *data.FileEvent
}

func (grpcInEndpoint *fileEventInEndPoint) HandleFileEvent(stream grpcservice.FileMonitoringActorSystemService_HandleFileEventServer) error {
	for {
		fileEvent, err := stream.Recv()
		if err != nil {
			log.Printf("Error reading from grpc event stream: %s", err)
			close(grpcInEndpoint.EventInChan)
			return nil
		}	
		grpcInEndpoint.EventInChan <- fileEvent
	}
}

func StartListener() chan *data.FileEvent {
	eventInChan := make(chan *data.FileEvent)
	grpcInEndpoint := fileEventInEndPoint{eventInChan}
	grpcServer := grpc.NewServer()
	
	grpcservice.RegisterFileMonitoringActorSystemServiceServer(grpcServer, &grpcInEndpoint)
	
	host,_ := os.Hostname()
	hostPortStr,exists := os.LookupEnv("ACTOR_GRPC_LISTENER_PORT")
	
	if !exists {
		log.Printf("ACTOR_GRPC_LISTENER_PORT property not set. Using 3000 as default grpc port")
		hostPortStr = "3000";
	}
	
	hostPort := fmt.Sprintf("%s:%s", utils.GetIpAddress(host),hostPortStr)
	
	log.Printf("Starting grpc listener on %s", hostPort)
	lis, err := net.Listen("tcp", hostPort)
	
	if err != nil {
		log.Fatal("Problem creating grpc listener:", err)
		return nil
	}
	
	go func() {
		log.Printf("Starting listener for grpc events.")
		grpcServer.Serve(lis)
		log.Printf("Started listener for grpc events.")
	}()
	
	return eventInChan
}
	
