package main

import (
	"github.com/saichler/habitat/service"
	"github.com/sirupsen/logrus"
)

func main() {
	s,err:=service.NewServiceManager()
	if err!=nil {
		logrus.Error("Failed to load habitat",err)
		return
	}

	s.Habitat().Uplink("192.168.86.29")
	//fmt.Println(uplinkHID.String())

	s.WaitForShutdown()
}
