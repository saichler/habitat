package service

import (
	. "github.com/saichler/habitat/golang/habitat"
	. "github.com/saichler/utils/golang"
	"sync"
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

type RepetitiveMessages struct {
	messages []*RepetitiveMessageEntry
	lock *sync.Cond
	timestamp int64
}

type RepetitiveMessageEntry struct {
	message   *Message
	interval  int64
	timestamp int64
}

type ServiceMessageHandler interface {
	Type() uint16
	HandleMessage(*ServiceManager,Service,*Message)
}

func NewRepetitiveMessages(srm *ServiceManager) *RepetitiveMessages {
	rm:=&RepetitiveMessages{}
	rm.lock = sync.NewCond(&sync.Mutex{})
	rm.messages=make([]*RepetitiveMessageEntry,0)
	go rm.repetitiveMessageSending(srm)
	return rm
}

func (rm *RepetitiveMessages) RegisterRepetitive(msg *Message,interval int64) {
	rm.lock.L.Lock()
	defer rm.lock.L.Unlock()
	e:=&RepetitiveMessageEntry{}
	e.message = msg
	e.interval = interval
	e.timestamp = time.Now().Unix()
	rm.messages = append(rm.messages,e)
}

func (rm *RepetitiveMessages) UnRegisterRepetitive(cid uint16) {
	rm.lock.L.Lock()
	defer rm.lock.L.Unlock()
	toRemove:=-1
	for i,v:=range rm.messages {
		if v.message.Source.ComponentID()==cid {
			toRemove = i
		}
	}
	tmp:=rm.messages[0:toRemove]
	tmp=append(tmp,rm.messages[toRemove+1:]...)
	rm.messages=tmp
}

func (rm *RepetitiveMessages) repetitiveMessageSending(srm *ServiceManager) {
	for;srm.habitat.Running(); {
		rm.lock.L.Lock()
		now:=time.Now().Unix()
		for _,ent:=range rm.messages {
			if now-ent.timestamp >ent.interval {
				srm.Send(ent.message)
				ent.timestamp = now
			}
		}
		rm.lock.L.Unlock()
		time.Sleep(time.Second*5)
	}
}

func (sh *ServiceHabitat) sendStartService() {
	Info("Sending Start Service For: "+sh.service.Name())
	msg:=sh.serviceManager.NewMessage(sh.service.ServiceID(),sh.service.ServiceID(),sh.service.ServiceID(), Message_Type_Service_START,[]byte(sh.service.Name()))
	sh.serviceManager.Send(msg)
}

func (sh *ServiceHabitat) repetitiveServicePing(rm *RepetitiveMessages) {
	Info("Adding repetitive ping for service: "+sh.service.Name())
	source:=NewServiceID(sh.serviceManager.habitat.HID(),sh.service.ServiceID().ComponentID(),sh.service.ServiceID().Topic())
	dest:=NewServiceID(PUBLISH_HID,sh.service.ServiceID().ComponentID(),sh.service.ServiceID().Topic())
	msg := sh.serviceManager.NewMessage(source, dest, source, Message_Type_Service_Ping, []byte(sh.service.Name()))
	rm.RegisterRepetitive(msg,10)
}
