package habitat

import (
	. "github.com/saichler/utils/golang"
	"sync"
)

type Packet struct {
	Source *HabitatID
	Dest   *HabitatID
	MID    uint32
	PID    uint32
	M      bool
	P      uint16
	Data []byte
}

var mba = NewByteArray()
var mbaMtx = &sync.Mutex{}

func (p *Packet) Marshal() []byte {
	mba.Reset()
	p.Source.Marshal(mba)
	p.Dest.Marshal(mba)
	mba.AddUInt32(p.MID)
	mba.AddUInt32(p.PID)
	mba.AddBool(p.M)
	mba.AddUInt16(p.P)
	mba.AddByteArray(encrypt(p.Data))
	return mba.Data()
}

func unmarshalPacketHeader(data []byte) (*HabitatID,*HabitatID,*ByteArray) {
	ba:=NewByteArrayWithData(data,0)
	source:=&HabitatID{}
	dest:=&HabitatID{}
	source.Unmarshal(ba)
	dest.Unmarshal(ba)
	return source,dest,ba
}

func (p *Packet) UnmarshalAll(source,dest *HabitatID,ba *ByteArray) {
	p.Source = source
	p.Dest = dest
	p.MID =ba.GetUInt32()
	p.PID=ba.GetUInt32()
	p.M=ba.GetBool()
	p.P=ba.GetUInt16()
	p.Data=decrypt(ba.GetByteArray())
}