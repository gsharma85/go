package config

import (

)

type ActorConfig struct {
	Name string
	Address string
	CompleteBy string
}

type ActorSystemConfig struct {
	Name string
	Address string
	ActorConfigs []ActorConfig
	EventNames []string
}