package service

import (
	"errors"
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	log "github.com/sirupsen/logrus"
	"plugin"
	"strconv"
	"sync"
)

type ServiceManager struct {
	habitat *Habitat
	services map[uint16]*ServiceHabitat
	lock *sync.Cond
}

type ServiceHabitat struct {
	serviceManager *ServiceManager
	serviceHandlers map[uint16]ServiceMessageHandler
	service Service
	inbox *Queue
}

func NewServiceManager() (*ServiceManager,error) {
	sm:=&ServiceManager{}
	habitat,err:=NewHabitat(sm)
	if err!=nil {
		log.Error("Failed to instantiate a Habitat")
		return nil,err
	}
	sm.services = make(map[uint16]*ServiceHabitat)
	sm.habitat = habitat
	sm.lock = sync.NewCond(&sync.Mutex{})
	return sm,nil
}

func (sm *ServiceManager) HID() *HID {
	return sm.habitat.HID()
}

func (sm *ServiceManager) InstallService(libraryPath string) error {
	servicePlugin, e := plugin.Open(libraryPath)
	if e!=nil {
		log.Error("Unable to load serivce plugin:",e)
		return e
	}
	svr,e:=servicePlugin.Lookup("ServiceInstance")
	if e!=nil {
		log.Error("Unable to find ServiceInstance in the library "+libraryPath)
		log.Error("Make sure you have: var ServiceInstance Service = &<your service struct>{}")
		return e
	}

	servicePtr,ok:=svr.(*Service)
	if !ok {
		msg:="ServiceInstance is not of type Service, please check that it implements Service and that ServiceInstance is a pointer."
		log.Error(msg)
		return errors.New(msg)
	}
	service:=*servicePtr

	log.Info("Service "+service.Name()+" was installed successfuly.")

	sm.AddService(service)

	return nil
}

func (sm *ServiceManager) AddService(service Service){
	sh:=&ServiceHabitat{}
	sh.service = service
	sh.service.SetManager(sm)
	sh.serviceManager = sm
	sh.inbox=NewQueue()
	sh.serviceHandlers = make(map[uint16]ServiceMessageHandler)
	for _,v:=range service.ServiceMessageHandlers() {
		sh.serviceHandlers[v.Type()]=v
	}
	sm.lock.L.Lock()
	sm.services[service.SID()]=sh
	sm.lock.L.Unlock()

	go sh.processNextMessage()
	sh.sendStartServiceMulticast()
}

func (sm *ServiceManager) getServiceHabitats() map[uint16]*ServiceHabitat {
	sm.lock.L.Lock()
	defer sm.lock.L.Unlock()
	services:=make(map[uint16]*ServiceHabitat)
	for k,v:=range sm.services {
		services[k]=v
	}
	return services
}

func (sm *ServiceManager) getServiceHabitat(message *Message) *ServiceHabitat {
	sm.lock.L.Lock()
	defer sm.lock.L.Unlock()
	return sm.services[message.Dest.CID]
}

func (sm *ServiceManager) HandleMessage(habitat *Habitat, message *Message){
	if message.IsMulticast() {
		shs:=sm.getServiceHabitats()
		for _,v:=range shs {
			v.inbox.Push(message)
		}
	} else {
		sh:=sm.getServiceHabitat(message)
		sh.inbox.Push(message)
	}
}

func (sm *ServiceManager) Shutdown() {
	sm.habitat.Shutdown()
}

func (sm *ServiceManager) WaitForShutdown(){
	sm.habitat.WaitForShutdown()
}

func (sm *ServiceManager) NewMessage(source, dest,origin *SID, ptype uint16,data []byte) *Message {
	return sm.habitat.NewMessage(source,dest,origin,ptype,data)
}

func (sm *ServiceManager) Send(message *Message) {
	sm.habitat.Send(message)
}

func (sm *ServiceManager) Source(s Service) *SID {
	return NewSID(sm.HID(),s.SID())
}

func (sm *ServiceManager) CreateAndReply(s Service,r *Message, ptype uint16,data []byte) {
	msg:=sm.NewMessage(sm.Source(s),r.Source,sm.Source(s),ptype,data)
	sm.Send(msg)
}

func(sm *ServiceManager) CreateAndSend(s Service,dest *SID,ptype uint16,data []byte) {
	msg := sm.NewMessage(sm.Source(s), dest, sm.Source(s), ptype, data)
	sm.Send(msg)
}

func (sh *ServiceHabitat) processNextMessage() {
	for ;sh.serviceManager.habitat.Running(); {
		message := sh.inbox.Pop().(*Message)
		msgHandler:=sh.serviceHandlers[message.Type]
		if msgHandler==nil {
			log.Error("There is no service message handler for message type:"+strconv.Itoa(int(message.Type)))
		}else {
			msgHandler.HandleMessage(sh.serviceManager,sh.service, message)
		}
	}
	log.Info("Service "+sh.service.Name()+" has shutdown")
}

func (svm *ServiceManager) Habitat() *Habitat {
	return svm.habitat
}