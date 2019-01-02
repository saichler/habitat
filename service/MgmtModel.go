package service

import (
	. "github.com/saichler/habitat"
	"strconv"
	"time"
	log "github.com/sirupsen/logrus"
)

type MgmtModel struct {
	Habitats map[HID]*HabitatInfo
}

type HabitatInfo struct {
	hid *HID
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
	hi.hid = hid
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
	}

	if si.LastPing==0 || time.Now().Unix()-si.LastPing>30 {
		log.Info("Service Manager Discovered Service ID:" + strconv.Itoa(int(sid)) + " Name:" + name + " in " + hi.hid.String())
	}

	si.LastPing = time.Now().Unix()
}