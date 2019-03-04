package service

import (
	. "github.com/saichler/habitat/golang/habitat"
	. "github.com/saichler/utils/golang"
)

type StartMgmtHandler struct {
}


func (h *StartMgmtHandler) Type() uint16 {
	return Message_Type_Service_START
}

func (h *StartMgmtHandler) HandleMessage(svm *ServiceManager,service Service,m *Message) {
	Info("Management Service for "+svm.HID().String()+" Is starting...")
	mgmtService:=service.(*MgmtService)
	mdl:=NewMgmtModel(svm.habitat.HID())
	mgmtService.Model = mdl
	hi:=mdl.AddHabitatInfo(svm.HID())
	hi.PutService(service.ServiceID().ComponentID(),service.ServiceID().Topic(),service.Name())
	Info("Management Service for "+svm.HID().String()+" Started!")
}
