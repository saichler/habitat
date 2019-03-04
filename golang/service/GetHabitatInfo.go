package service

import (
	. "github.com/saichler/habitat/golang/habitat"
)

type GetHabitatInfo struct {
}


func (h *GetHabitatInfo) Type() uint16 {
	return Mgmt_Type_Get_Info
}

func (h *GetHabitatInfo) HandleMessage(svm *ServiceManager,service Service,m *Message) {
}
