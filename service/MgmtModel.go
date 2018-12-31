package service

import (
	. "github.com/saichler/habitat"
	"github.com/sirupsen/logrus"
	"strconv"
)

type MgmtModel struct {
	Habitats map[HID]*HabitatInfo
}

type HabitatInfo struct {
	Services map[uint16]string
}

func NewMgmtModel () *MgmtModel {
	mgmt:=&MgmtModel{}
	mgmt.Habitats = make(map[HID]*HabitatInfo)
	return mgmt
}

func (m *MgmtModel) AddHabitatInfo(hid *HID) *HabitatInfo {
	hi:=&HabitatInfo{}
	hi.Services = make(map[uint16]string)
	m.Habitats[*hid]=hi
	return hi
}

func (m *MgmtModel) GetHabitatInfo(hid *HID) *HabitatInfo {
	return m.Habitats[*hid]
}

func (hi *HabitatInfo) AddService(sid uint16,name string) {
	hi.Services[sid]=name
	logrus.Info("Service Manager Added Service ID:"+strconv.Itoa(int(sid))+" Name:"+name)
}