package habitat

import (
	. "github.com/saichler/utils/golang"
)

type Packet struct {
	Source *HabitatID
	Dest   *HabitatID
	MultiPart bool
	Persisted bool
	Priority      int //uint6 if existed
	MID    uint32
	PID    uint32
	Data []byte
}

func (p *Packet) Marshal() []byte {
	bs :=NewByteSlice()
	p.Source.Marshal(bs)
	p.Dest.Marshal(bs)
	mpp:=Encode2BoolAndUInt6(p.MultiPart,p.Persisted,p.Priority)
	bs.AddByte(mpp)
	bs.AddUInt32(p.MID)
	bs.AddUInt32(p.PID)
	bs.AddByteSlice(encrypt(p.Data))
	return bs.Data()
}

func unmarshalPacketHeader(data []byte) (*HabitatID,*HabitatID,bool,bool,int,*ByteSlice) {
	ba:=NewByteSliceWithData(data,0)
	source:=&HabitatID{}
	dest:=&HabitatID{}
	source.Unmarshal(ba)
	dest.Unmarshal(ba)
	mpp:=ba.GetByte()
	m,prs,pri:=Decode2BoolAndUInt6(mpp)
	return source,dest,m,prs,pri,ba
}

func (p *Packet) UnmarshalAll(source,dest *HabitatID,m,prs bool,pri int,ba *ByteSlice) {
	p.Source = source
	p.Dest = dest
	p.MultiPart = m
	p.Persisted = prs
	p.Priority = pri
	p.MID =ba.GetUInt32()
	p.PID=ba.GetUInt32()
	p.Data=decrypt(ba.GetByteSlice())
}

func GetPriority(data []byte) int {
	p:=0
	//if len(data)>32{
		_, _, p = Decode2BoolAndUInt6(data[32])
	//}
	return p
}