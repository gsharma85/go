package grpc

import (
	"github.com/gsharma85/go/dataflow/pkg/data"
	grpcservice "github.com/gsharma85/go/dataflow/pkg/grpc"
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
	
	hostPort := fmt.Sprintf("%s:%s", getIpAddress(host),os.Getenv("ACTOR_GRPC_LISTENER_PORT"))
	
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

func getIpAddress(host string) string {
	addrs, _ := net.LookupIP(host)
	for _, addr := range addrs {
	    if ipv4 := addr.To4(); ipv4 != nil {
	        return ipv4.String()
	    }      
	}
	return ""
}	
