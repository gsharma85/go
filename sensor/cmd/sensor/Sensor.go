package main

import (
	. "github.com/gsharma85/go/sensor/internal/core"
	. "github.com/gsharma85/go/sensor/internal/data"
	_ "github.com/gsharma85/go/sensor/pkg/grpc"
)

func main() {
	stopChan := make(chan struct{})
	
	senseConfig := SenseConfig{stopChan, "C:\\gotest"}
	
	StartSensorPipeline("file", "console", senseConfig)
	
	<- stopChan
}
