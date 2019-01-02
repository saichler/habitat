package service

import (
	. "github.com/saichler/habitat"
	"github.com/sirupsen/logrus"
)

type StartMgmtHandler struct {
}


func (h *StartMgmtHandler) Type() uint16 {
	return Message_Type_Service_START
}

func (h *StartMgmtHandler) HandleMessage(svm *ServiceManager,service Service,m *Message) {
	logrus.Info("***** Management Service for "+svm.HID().String()+" Is starting... *****")
	mgmtService:=service.(*MgmtService)
	mdl:=NewMgmtModel(svm.habitat.HID())
	mgmtService.Model = mdl
	hi:=mdl.AddHabitatInfo(svm.HID())
	hi.PutService(service.SID(),service.Name())
	logrus.Info("***** Management Service for "+svm.HID().String()+" Started! *****")
}
