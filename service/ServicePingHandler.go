package service

import (
	. "github.com/saichler/habitat"
)

type ServicePingHandler struct {
}


func (h *ServicePingHandler) Type() uint16 {
	return Message_Type_Service_Ping
}

func (h *ServicePingHandler) HandleMessage(svm *ServiceManager,service Service,m *Message) {
	svm.ServicePing(m.Source,string(m.Data))
}
