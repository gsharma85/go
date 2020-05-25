package data

import (
	"time"
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
}