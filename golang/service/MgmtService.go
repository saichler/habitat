package service

import . "github.com/saichler/habitat/golang/habitat"

const (
	MANAGEMENT_SERVICE_TOPIC="Management Service Topic"
	MANAGEMENT_ID uint16 = 10
)

type MgmtService struct {
	svm *ServiceManager
	sid *ServiceID
	Model *MgmtModel
}


func (s *MgmtService) Name() string {
	return "Habitat Management Service"
}

func (s *MgmtService) ServiceID() *ServiceID {
	return s.sid
}

func (s *MgmtService) ServiceManager() *ServiceManager {
	return s.svm
}

func (s *MgmtService) Init(svm *ServiceManager,cid uint16) {
	s.svm = svm
	s.sid = NewServiceID(svm.habitat.HID(),cid,MANAGEMENT_SERVICE_TOPIC)
}

func (s *MgmtService) ServiceMessageHandlers()[]ServiceMessageHandler {
	return []ServiceMessageHandler {
		&StartMgmtHandler{},
		&ServicePingHandler{}}
}

func (s *MgmtService) UnreachableMessageHandlers()[]ServiceMessageHandler {
	return s.ServiceMessageHandlers()
}