package service

import (
	. "github.com/saichler/habitat"
	"sync"
)

type ServiceManager struct {
	habitat *Habitat
	services map[SID]Service
	lock *sync.Cond
}

func NewServiceManager(habitat *Habitat) *ServiceManager {
	sm:=&ServiceManager{}
	sm.services = make(map[SID]Service)
	sm.habitat = habitat
	sm.lock = sync.NewCond(&sync.Mutex{})
	return sm
}

func (sm *ServiceManager) AddService(service Service){
	sm.lock.L.Lock()
	defer sm.lock.L.Unlock()
	sm.services[*service.SID()]=service
}

func (sm *ServiceManager) HandleMessage(habitat *Habitat, message *Message){

}