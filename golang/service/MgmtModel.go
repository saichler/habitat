package service

import (
	. "github.com/saichler/habitat/golang/habitat"
	. "github.com/saichler/utils/golang"
	"strconv"
	"time"
)

type MgmtModel struct {
	myHid *HabitatID
	Habitats *ConcurrentMap
}

type HabitatInfo struct {
	hid *HabitatID
	Services *ConcurrentMap
}

type ServiceInfo struct {
	Cid uint16
	Topic string
	Name string
	LastPing int64
}

func NewMgmtModel (myHid *HabitatID) *MgmtModel {
	mgmt:=&MgmtModel{}
	mgmt.myHid = myHid
	mgmt.Habitats = NewConcurrentMap()
	return mgmt
}

func (m *MgmtModel) AddHabitatInfo(hid *HabitatID) *HabitatInfo {
	hi:=&HabitatInfo{}
	hi.Services = NewConcurrentMap()
	hi.hid = hid
	m.Habitats.Put(*hid,hi)
	return hi
}

func (m *MgmtModel) GetHabitatInfo(hid *HabitatID) *HabitatInfo {
	hi,_:=m.Habitats.Get(*hid)
	if hi==nil {
		hi = m.AddHabitatInfo(hid)
	}
	return hi.(*HabitatInfo)
}

func (hi *HabitatInfo) PutService(componentId uint16,topic,name string) {
	existing,_:=hi.Services.Get(componentId)
	var si *ServiceInfo
	if existing==nil {
		si=&ServiceInfo{}
		si.Name = name
		si.Topic = topic
		si.Cid = componentId
		hi.Services.Put(componentId,si)
	} else {
		si=existing.(*ServiceInfo)
	}

	if si.LastPing==0 || time.Now().Unix()-si.LastPing>30 {
		Info("Service Manager Discovered Service ID:" + strconv.Itoa(int(componentId)) + " Name:" + name + " in " + hi.hid.String())
	}

	si.LastPing = time.Now().Unix()
}

func (model *MgmtModel) GetAllServicesOfType(topic string) []*ServiceID {
	result:=make([]*ServiceID,0)
	allHabitats:=model.Habitats.GetMap()
	for k,v:=range allHabitats {
		key:=k.(HabitatID)
		value:=v.(*HabitatInfo)
		if !key.Equal(model.myHid) {
			allServices:=value.Services.GetMap()
			for _,s:=range allServices {
				si:=s.(*ServiceInfo)
				if si.Topic==topic {
					hid:=&HabitatID{}
					hid.UuidM = key.UuidM
					hid.UuidL = key.UuidL
					result = append(result,NewServiceID(hid,si.Cid,si.Topic))
				}
			}
		}
	}
	return result
}