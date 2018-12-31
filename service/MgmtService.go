package service

import (
	"strconv"
)

type MgmtService struct {
	svm *ServiceManager
	Model *MgmtModel
}

func (s *MgmtService) SID() uint16 {
	return 2
}

func (s *MgmtService) Name() string {
	return "Habitat Management Service ID="+strconv.Itoa(int(s.SID()))
}

func (s *MgmtService) ServiceMessageHandlers()[]ServiceMessageHandler {
	return []ServiceMessageHandler {
		&StartMgmtHandler{}}
}

func(s *MgmtService) SetManager(svm *ServiceManager) {
	s.svm = svm
}

func(s *MgmtService) GetManager() *ServiceManager{
	return s.svm
}