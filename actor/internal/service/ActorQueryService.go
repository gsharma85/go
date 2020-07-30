package service

import (
	"context"
	"net/http"
	"log"
	"os"
	"fmt"
	"encoding/json"
	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gsharma85/go/actor/internal/actor"
	"github.com/gsharma85/go/actor/internal/utils"
)

type QueryActorAlertsService interface {
	Execute(actorSystemType string) []string
}

type queryActorAlertsService struct {

}

type QueryActorStatusService interface {
	Execute(actorSystemType string) []string
}

type queryActorStatusService struct {

}

func (service *queryActorAlertsService) Execute(actorSystemType string) []string{
	switch actorSystemType {
	case "file":
		return actor.GetFileActorAlerts()
	}
	return make([]string, 0)
}

func (service *queryActorStatusService) Execute(actorSystemType string) []string{
	switch actorSystemType {
	case "file":
		return actor.GetFileActorStatus()
	}
	return make([]string, 0)
}

type QueryActorRequest struct {
	ActorSystemType string "json:actorSystemType"
}

type QueryActorResponse struct {
	Alerts []string "json:alerts"
}

func makeQueryAlertsEndpoint(service QueryActorAlertsService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		queryActorRequest := request.(QueryActorRequest)
		alerts := service.Execute(queryActorRequest.ActorSystemType)
		return QueryActorResponse{alerts}, nil
	}
}

func makeQueryStatusEndpoint(service QueryActorStatusService) endpoint.Endpoint {
	return func(_ context.Context, request interface{}) (interface{}, error) {
		queryActorRequest := request.(QueryActorRequest)
		alerts := service.Execute(queryActorRequest.ActorSystemType)
		return QueryActorResponse{alerts}, nil
	}
}

func StartQueryActorServer() {
	serviceAlert := queryActorAlertsService{}
	serviceStatus := queryActorStatusService{}
	
	handlerAlerts := httptransport.NewServer(makeQueryAlertsEndpoint(&serviceAlert), decodeQueryRequest, encodeQueryResponse)
	
	handlerStatus := httptransport.NewServer(makeQueryStatusEndpoint(&serviceStatus), decodeQueryRequest, encodeQueryResponse)
	
	http.Handle("/alerts", handlerAlerts)
	http.Handle("/status", handlerStatus)
	
	host,_ := os.Hostname()
	
	hostPort := fmt.Sprintf("%s:%s", utils.GetIpAddress(host),os.Getenv("ACTOR_HTTP_PORT"))
	
	log.Fatal(http.ListenAndServe(hostPort, nil))
}

func decodeQueryRequest(_ context.Context, httprequest *http.Request) (interface{}, error){
	var request QueryActorRequest
	if err := json.NewDecoder(httprequest.Body).Decode(&request); err != nil {
		return nil, err
	}
	return request, nil
}

func encodeQueryResponse(_ context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}

