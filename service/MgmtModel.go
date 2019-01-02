package service

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	log "github.com/sirupsen/logrus"
	"strconv"
	"time"
)

type MgmtModel struct {
	myHid *HID
	Habitats *ConcurrentMap
}

type HabitatInfo struct {
	hid *HID
	Services *ConcurrentMap
}

type ServiceInfo struct {
	SID uint16
	Name string
	LastPing int64
}

func NewMgmtModel (myHid *HID) *MgmtModel {
	mgmt:=&MgmtModel{}
	mgmt.myHid = myHid
	mgmt.Habitats = NewConcurrentMap()
	return mgmt
}

func (m *MgmtModel) AddHabitatInfo(hid *HID) *HabitatInfo {
	hi:=&HabitatInfo{}
	hi.Services = NewConcurrentMap()
	hi.hid = hid
	m.Habitats.Put(*hid,hi)
	return hi
}

func (m *MgmtModel) GetHabitatInfo(hid *HID) *HabitatInfo {
	hi,_:=m.Habitats.Get(*hid)
	if hi==nil {
		hi = m.AddHabitatInfo(hid)
	}
	return hi.(*HabitatInfo)
}

func (hi *HabitatInfo) PutService(sid uint16,name string) {
	existing,_:=hi.Services.Get(sid)
	var si *ServiceInfo
	if existing==nil {
		si=&ServiceInfo{}
		si.Name = name
		si.SID = sid
		hi.Services.Put(sid,si)
	} else {
		si=existing.(*ServiceInfo)
	}

	if si.LastPing==0 || time.Now().Unix()-si.LastPing>30 {
		log.Info("Service Manager Discovered Service ID:" + strconv.Itoa(int(sid)) + " Name:" + name + " in " + hi.hid.String())
	}

	si.LastPing = time.Now().Unix()
}

func (model *MgmtModel) GetAllServicesOfType(t uint16) []*SID {
	result:=make([]*SID,0)
	allHabitats:=model.Habitats.GetMap()
	for k,v:=range allHabitats {
		key:=k.(HID)
		value:=v.(*HabitatInfo)
		if !key.Equal(model.myHid) {
			allServices:=value.Services.GetMap()
			for s,_:=range allServices {
				sid:=s.(uint16)
				if sid==t {
					hid:=&HID{}
					hid.UuidM = key.UuidM
					hid.UuidL = key.UuidL
					result = append(result,NewSID(hid,sid))
				}
			}
		}
	}
	return result
}