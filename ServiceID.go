package habitat

import (
	"bytes"
	. "github.com/saichler/utils/golang"
	"strconv"
)

type ServiceID struct {
	hid *HabitatID
	cid uint16
	topic string
}

func NewServiceID(hid *HabitatID,componentId uint16,topic string) *ServiceID {
	sid:=&ServiceID{}
	sid.hid = hid
	sid.cid = componentId
	sid.topic = topic
	return sid
}

func (sid *ServiceID) Marshal(ba *ByteArray) {
	sid.hid.Marshal(ba)
	ba.AddUInt16(sid.cid)
	ba.AddString(sid.topic)
}

func (sid *ServiceID) Unmarshal(ba *ByteSlice) {
	sid.hid = &HabitatID{}
	sid.hid.UuidM = ba.GetInt64()
	sid.hid.UuidL = ba.GetInt64()
	sid.cid = ba.GetUInt16()
	sid.topic = ba.GetString()
}

func (sid *ServiceID) IsPublish() bool {
	return sid.hid.IsPublish()
}

func (sid *ServiceID)  String() string {
	buff:=bytes.Buffer{}
	buff.WriteString(sid.hid.String())
	buff.WriteString("[")
	buff.WriteString("CID=")
	buff.WriteString(strconv.Itoa(int(sid.cid)))
	buff.WriteString(",Topic=")
	buff.WriteString(sid.topic)
	buff.WriteString("]")
	return buff.String()
}

func (sid *ServiceID) Hid() *HabitatID {
	return sid.hid
}

func (sid *ServiceID) ComponentID() uint16 {
	return sid.cid
}

func (sid *ServiceID) Topic() string {
	return sid.topic
}