package main

import (
	. "github.com/saichler/habitat/service"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
)

func main() {
	s,err:=NewServiceManager()
	if err!=nil {
		logrus.Error("Failed to load habitat",err)
		return
	}
	args := os.Args[1:]

	for _,arg:=range args {
		if strings.Contains(arg,".so") {
			s.InstallService(arg)
		} else {
			s.Habitat().Uplink(arg)
		}
	}
	s.WaitForShutdown()
}
