package service

import (
	. "github.com/saichler/habitat"
	log "github.com/sirupsen/logrus"
	"time"
)

type Service interface {
	Name() string
	SID() uint16
	ServiceMessageHandlers()[]ServiceMessageHandler
	SetManager(*ServiceManager)
	GetManager() *ServiceManager
}

type ServiceMessageHandler interface {
	Type() uint16
	HandleMessage(*ServiceManager,Service,*Message)
}


func (sh *ServiceHabitat) sendStartService() {
	source:=NewSID(sh.serviceManager.habitat.HID(),sh.service.SID())
	msg:=sh.serviceManager.NewMessage(source,source,source, Message_Type_Service_START,[]byte(sh.service.Name()))
	sh.serviceManager.Send(msg)
}

func (sh *ServiceHabitat) sendServicePingMulticast() {
	log.Info("Adding multicast ping for service: "+sh.service.Name())
	time.Sleep(time.Second)
	lastSent:=int64(0)
	for ;sh.serviceManager.habitat.Running(); {
		if time.Now().Unix()-lastSent>15 {
			source := NewSID(sh.serviceManager.habitat.HID(), sh.service.SID())
			dest := NewSID(NewMulticastHID(sh.service.SID()), sh.service.SID())
			msg := sh.serviceManager.NewMessage(source, dest, source, Message_Type_Service_Ping, []byte(sh.service.Name()))
			sh.serviceManager.Send(msg)
			lastSent=time.Now().Unix()
		}
	}
}
