package data

import (
	"time"
	"github.com/gsharma85/go/dataflow/pkg/data"
)
type SenseStatus uint32

const (
	SUCCESS SenseStatus = iota
	FAILURE
)

type SenseEvent struct {
	Key string
	Status SenseStatus
	Action string
	Time time.Time
}

type SenseConfig struct {
	StopSignal chan struct{}
	DirName string
	DeDupInterval time.Duration
	DeDupKeyGen func(senseEvent SenseEvent) (groupByKey string, eventKey string)
	Logfile string
	ActorSystemConfigData *data.ActorSystemConfig
}