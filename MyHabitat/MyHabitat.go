package main

import (
	"github.com/saichler/habitat/service"
	"github.com/sirupsen/logrus"
	"os"
)

func main() {
	s,err:=service.NewServiceManager()
	if err!=nil {
		logrus.Error("Failed to load habitat",err)
		return
	}
	args := os.Args[1:]

	for _,uplink:=range args {
		s.Habitat().Uplink(uplink)
	}
	
	s.WaitForShutdown()
}
