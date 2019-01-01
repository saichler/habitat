package service

import (
	. "github.com/saichler/habitat"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type MgmtModel struct {
	Habitats map[HID]*HabitatInfo
}

type HabitatInfo struct {
	Services map[uint16]*ServiceInfo
}

type ServiceInfo struct {
	SID uint16
	Name string
	LastPing int64
}

func NewMgmtModel () *MgmtModel {
	mgmt:=&MgmtModel{}
	mgmt.Habitats = make(map[HID]*HabitatInfo)
	return mgmt
}

func (m *MgmtModel) AddHabitatInfo(hid *HID) *HabitatInfo {
	hi:=&HabitatInfo{}
	hi.Services = make(map[uint16]*ServiceInfo)
	m.Habitats[*hid]=hi
	return hi
}

func (m *MgmtModel) GetHabitatInfo(hid *HID) *HabitatInfo {
	hi:=m.Habitats[*hid]
	if hi==nil {
		hi = m.AddHabitatInfo(hid)
	}
	return hi
}

func (hi *HabitatInfo) PutService(sid uint16,name string) {
	si,ok:=hi.Services[sid]
	if !ok {
		si=&ServiceInfo{}
		si.Name = name
		si.SID = sid
		hi.Services[sid]=si
		log.Info("Service Manager Discovered Service ID:"+strconv.Itoa(int(sid))+" Name:"+name)
	}
	si.LastPing = time.Now().Unix()
}