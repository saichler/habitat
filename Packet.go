package habitat

import (
	. "github.com/saichler/utils/golang"
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

func (p *Packet) Marshal() []byte {
	bs :=NewByteSlice()
	p.Source.Marshal(bs)
	p.Dest.Marshal(bs)
	bs.AddUInt32(p.MID)
	bs.AddUInt32(p.PID)
	bs.AddBool(p.M)
	bs.AddUInt16(p.P)
	bs.AddByteSlice(encrypt(p.Data))
	return bs.Data()
}

func unmarshalPacketHeader(data []byte) (*HabitatID,*HabitatID,*ByteSlice) {
	ba:=NewByteSliceWithData(data,0)
	source:=&HabitatID{}
	dest:=&HabitatID{}
	source.Unmarshal(ba)
	dest.Unmarshal(ba)
	return source,dest,ba
}

func (p *Packet) UnmarshalAll(source,dest *HabitatID,ba *ByteSlice) {
	p.Source = source
	p.Dest = dest
	p.MID =ba.GetUInt32()
	p.PID=ba.GetUInt32()
	p.M=ba.GetBool()
	p.P=ba.GetUInt16()
	p.Data=decrypt(ba.GetByteSlice())
}