package model

import (
	"github.com/saichler/habitat/golang/service"
)

type Console struct {
	ServiceManager *service.ServiceManager
	Commands       map[string]Command
	Running        bool
}

type ConsoleConfig struct {
	Host string `PK:"host:0"`
}

const (
	CONFIG_FILE = "./resources/console-config.yaml"
)
