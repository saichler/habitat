package service

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	"time"
)

type Service interface {
	Name() string
	ServiceID() *ServiceID
	ServiceManager() *ServiceManager
	Init(*ServiceManager,uint16)
	ServiceMessageHandlers()[]ServiceMessageHandler
	UnreachableMessageHandlers()[]ServiceMessageHandler
}

type ServiceMessageHandler interface {
	Type() uint16
	HandleMessage(*ServiceManager,Service,*Message)
}

func (sh *ServiceHabitat) sendStartService() {
	Info("Sending Start Service For: "+sh.service.Name())
	msg:=sh.serviceManager.NewMessage(sh.service.ServiceID(),sh.service.ServiceID(),sh.service.ServiceID(), Message_Type_Service_START,[]byte(sh.service.Name()))
	sh.serviceManager.Send(msg)
}

func (sh *ServiceHabitat) repetitiveServicePing() {
	Info("Adding repetitive ping for service: "+sh.service.Name())
	time.Sleep(time.Second)
	lastSent:=int64(0)
	for ;sh.serviceManager.habitat.Running(); {
		if time.Now().Unix()-lastSent>5 {
			source:=NewServiceID(sh.serviceManager.habitat.HID(),sh.service.ServiceID().ComponentID(),sh.service.ServiceID().Topic())
			dest:=NewServiceID(PUBLISH_HID,sh.service.ServiceID().ComponentID(),sh.service.ServiceID().Topic())
			msg := sh.serviceManager.NewMessage(source, dest, source, Message_Type_Service_Ping, []byte(sh.service.Name()))
			sh.serviceManager.Send(msg)
			lastSent=time.Now().Unix()
		}
		time.Sleep(time.Second)
	}
}
