package service

import (
	. "github.com/saichler/habitat"
	"sync"
)

type ServiceManager struct {
	habitat *Habitat
	services map[HID]Service
	lock *sync.Cond
}

func NewServiceManager(habitat *Habitat) *ServiceManager {
	sm:=&ServiceManager{}
	sm.habitat = habitat
	sm.lock = sync.NewCond(&sync.Mutex{})
	return sm
}

func (sm *ServiceManager) AddService(service Service){
	sm.lock.L.Lock()
	defer sm.lock.L.Unlock()
	sm.services[*service.HID()]=service
}

func (sm *ServiceManager) HandleMessage(habitat *Habitat, message *Message){

}