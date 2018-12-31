package service

import . "github.com/saichler/habitat"
import log "github.com/sirupsen/logrus"
var EMPTY=make([]byte,0)

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
	msg:=sh.serviceManager.NewMessage(source,source,source, Message_Type_Service_START,EMPTY)
	sh.serviceManager.Send(msg)
}

func (sh *ServiceHabitat) sendServiceStartedMulticast() {
	source:=NewSID(sh.serviceManager.habitat.HID(),sh.service.SID())
	dest:=NewSID(NewMulticastHID(sh.service.SID()),sh.service.SID())
	msg:=sh.serviceManager.NewMessage(source,dest,source, Message_Type_Service_STARTED,EMPTY)
    log.Info("Sending Start Message to "+sh.service.Name())
	sh.serviceManager.Send(msg)
}
