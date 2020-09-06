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
	grpcServer *grpc.Server
}

func (grpcInEndpoint *fileEventInEndPoint) HandleFileEvent(stream grpcservice.FileMonitoringActorSystemService_HandleFileEventServer) error {
	for {
		fileEvent, err := stream.Recv()
		if err != nil {
			log.Printf("Error reading from grpc event stream: %s", err)
			grpcInEndpoint.restartListener()
			return err
		} else if fileEvent != nil {
			grpcInEndpoint.EventInChan <- fileEvent
		}		
	}
}

func (grpcInEndpoint *fileEventInEndPoint) restartListener() {
	grpcservice.RegisterFileMonitoringActorSystemServiceServer(grpcInEndpoint.grpcServer, grpcInEndpoint)
}

func StartListener() chan *data.FileEvent {
	eventInChan := make(chan *data.FileEvent)
	grpcServer := grpc.NewServer()
	
	grpcInEndpoint := fileEventInEndPoint{eventInChan, grpcServer}
	
	grpcservice.RegisterFileMonitoringActorSystemServiceServer(grpcServer, &grpcInEndpoint)
	
	host,_ := os.Hostname()
	
	hostPortStr,exists := os.LookupEnv("ACTOR_GRPC_LISTENER_PORT")
	
	if !exists {
		log.Printf("ACTOR_GRPC_LISTENER_PORT property not set. Using 3000 as default grpc port")
		hostPortStr = "3000";
	}
	
	_, exists = os.LookupEnv("ACTOR_GRPC_LISTENER_LOCAL_HOST")
	
	var hostIp string
	if exists {
		hostIp = "127.0.0.1"
	} else {
		hostIp = utils.GetIpAddress(host)
	}
	
	log.Printf("Using %s interace to listen on grpc stream.", hostIp)
	
	hostPort := fmt.Sprintf("%s:%s", hostIp ,hostPortStr)
	
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
	
