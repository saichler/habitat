package service

import (
	"errors"
	. "github.com/saichler/habitat/golang/habitat"
	. "github.com/saichler/utils/golang"
	"plugin"
	"strconv"
	"sync"
)

type ServiceManager struct {
	habitat *Habitat
	services *ConcurrentMap
	lock *sync.Cond
	topics map[string][]uint16
	nextComponentID uint16
	repetitive *RepetitiveMessages
}

type ServiceHabitat struct {
	serviceManager *ServiceManager
	serviceHandlers map[uint16]ServiceMessageHandler
	unreachableHandlers map[uint16]ServiceMessageHandler
	service Service
	inbox *PriorityQueue
}

func NewServiceManager() (*ServiceManager,error) {
	sm:=&ServiceManager{}
	habitat,err:=NewHabitat(sm)
	if err!=nil {
		Error("Failed to instantiate a Habitat")
		return nil,err
	}
	sm.services = NewConcurrentMap()
	sm.topics = make(map[string][]uint16)
	sm.nextComponentID = 10
	sm.repetitive = NewRepetitiveMessages(sm)

	sm.habitat = habitat
	sm.lock = sync.NewCond(&sync.Mutex{})

	mgmt:=&MgmtService{}
	sm.AddService(mgmt)

	return sm,nil
}

func (sm *ServiceManager) RegisterForTopic(service Service,topic string) {
	_,ok:=sm.topics[topic]
	if !ok {
		sm.topics[topic]=make([]uint16,0)
	}
	sm.topics[topic] = append(sm.topics[topic],service.ServiceID().ComponentID())
}

func (sm *ServiceManager) HID() *HabitatID {
	return sm.habitat.HID()
}

func (sm *ServiceManager) InstallService(libraryPath string) error {
	servicePlugin, e := plugin.Open(libraryPath)
	if e!=nil {
		Error("Unable to load serivce plugin:",e)
		return e
	}
	svr,e:=servicePlugin.Lookup("ServiceInstance")
	if e!=nil {
		Error("Unable to find ServiceInstance in the library "+libraryPath)
		Error("Make sure you have: var ServiceInstance Service = &<your service struct>{}")
		return e
	}

	servicePtr,ok:=svr.(*Service)
	if !ok {
		msg:="ServiceInstance is not of type Service, please check that it implements Service and that ServiceInstance is a pointer."
		Error(msg)
		return errors.New(msg)
	}
	service:=*servicePtr

	Info("Service "+service.Name()+" was installed successfuly.")

	sm.AddService(service)

	return nil
}

func (sm *ServiceManager) RemoveService(cid uint16) {
	sm.repetitive.UnRegisterRepetitive(cid)
	sm.services.Del(cid)
}

func (sm *ServiceManager) AddService(service Service){
	sm.lock.L.Lock()
	componentId:=sm.nextComponentID
	sm.nextComponentID++
	sm.lock.L.Unlock()

	service.Init(sm,componentId)

	sh:=&ServiceHabitat{}
	sh.service = service
	sh.serviceManager = sm
	sh.inbox=NewPriorityQueue()
	sh.inbox.SetName(service.Name())
	sh.serviceHandlers = make(map[uint16]ServiceMessageHandler)
	for _,v:=range service.ServiceMessageHandlers() {
		sh.serviceHandlers[v.Type()]=v
	}
	sh.unreachableHandlers = make(map[uint16]ServiceMessageHandler)
	for _,v:=range service.UnreachableMessageHandlers() {
		sh.unreachableHandlers[v.Type()]=v
	}

	sm.services.Put(service.ServiceID().ComponentID(),sh)

	go sh.processNextMessage()

	if service.ServiceID().Topic()!=MANAGEMENT_SERVICE_TOPIC {
		sm.getManagementService().Model.GetHabitatInfo(sm.HID()).PutService(service.ServiceID().ComponentID(),service.ServiceID().Topic(), service.Name())
		sh.sendStartService()
	} else {
		mh:= StartMgmtHandler{}
		mh.HandleMessage(sm,service,nil)
	}
	sh.repetitiveServicePing(sm.repetitive)
}

func (sm *ServiceManager) getManagementService() *MgmtService {
	sh,ok:=sm.services.Get(MANAGEMENT_ID)
	if !ok {
		panic("not ok")
	}
	return sh.(*ServiceHabitat).service.(*MgmtService)
}

func (sm *ServiceManager) getServiceHabitat(message *Message) *ServiceHabitat {
	cid:=message.Dest.ComponentID()
	if cid==0 {
		allHabitats:=sm.services.GetMap()
		for k,v:=range allHabitats {
			sh:=v.(*ServiceHabitat)
			if sh.service.ServiceID().Topic()==message.Dest.Topic() {
				cid = k.(uint16)
				break
			}
		}
	}
	result,_:=sm.services.Get(cid)
	return result.(*ServiceHabitat)
}

func (sm *ServiceManager) getServiceHabitatByID(cid uint16) *ServiceHabitat {
	result,_:=sm.services.Get(cid)
	return result.(*ServiceHabitat)
}

func (sm *ServiceManager) HandleUnreachable(habitat *Habitat, message *Message){

}

func (sm *ServiceManager) HandleMessage(habitat *Habitat, message *Message){
	if message.IsPublish() {
		if message.Type==Message_Type_Service_Ping {
			msh:=sm.getServiceHabitatByID(MANAGEMENT_ID)
			msh.inbox.Push(message,message.Priority)
			return
		}
		rg:=sm.topics[message.Dest.Topic()]
		if rg!=nil {
			for _,sid:=range rg {
				srv,_:=sm.services.Get(sid)
				service:=srv.(*ServiceHabitat)
				service.inbox.Push(message,message.Priority)
			}
		}
	} else {
		sh:=sm.getServiceHabitat(message)
		sh.inbox.Push(message,message.Priority)
	}
}

func (sm *ServiceManager) Shutdown() {
	sm.habitat.Shutdown()
	services:=sm.services.GetMap()
	for _,s:=range services {
		sh:=s.(*ServiceHabitat)
		sh.inbox.Shutdown()
	}
}

func (sm *ServiceManager) WaitForShutdown(){
	sm.habitat.WaitForShutdown()
}

func (sm *ServiceManager) NewMessage(source, dest,origin *ServiceID, ptype uint16,data []byte) *Message {
	return sm.habitat.NewMessage(source,dest,origin,ptype,data)
}

func (sm *ServiceManager) Send(message *Message) {
	sm.habitat.Send(message)
}

func (sm *ServiceManager) CreateAndReply(s Service,r *Message, ptype uint16,data []byte) {
	msg:=sm.NewMessage(s.ServiceID(),r.Source,s.ServiceID(),ptype,data)
	sm.Send(msg)
}

func(sm *ServiceManager) CreateAndSend(s Service,dest *ServiceID,ptype uint16,data []byte) {
	msg := sm.NewMessage(s.ServiceID(), dest, s.ServiceID(), ptype, data)
	sm.Send(msg)
}

func (sh *ServiceHabitat) processNextMessage() {
	for ;sh.serviceManager.habitat.Running(); {
		msg := sh.inbox.Pop()
		if msg==nil {
			break;
		}
		message:=msg.(*Message)
		if message.Unreachable {
			msgHandler := sh.unreachableHandlers[message.Type]
			if msgHandler == nil {
				Error("There is no service message unreachable handler for message type:" + strconv.Itoa(int(message.Type)))
			} else {
				message.Unreachable = true
				msgHandler.HandleMessage(sh.serviceManager, sh.service, message)
			}
		} else {
			msgHandler := sh.serviceHandlers[message.Type]
			if msgHandler == nil {
				Error("There is no service message handler for message type:" + strconv.Itoa(int(message.Type)))
			} else {
				msgHandler.HandleMessage(sh.serviceManager, sh.service, message)
			}
		}
	}
	Info("Service "+sh.service.Name()+" has shutdown")
}

func (svm *ServiceManager) Habitat() *Habitat {
	return svm.habitat
}

func (svm *ServiceManager) ServicePing(sid *ServiceID, name string) {
	svm.getManagementService().Model.GetHabitatInfo(sid.Hid()).PutService(sid.ComponentID(),sid.Topic(),name)
}

func (svm *ServiceManager) GetAllAdjacents(service Service) []*ServiceID {
	return svm.getManagementService().Model.GetAllServicesOfType(service.ServiceID().Topic())
}