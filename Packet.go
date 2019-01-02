package habitat

import . "github.com/saichler/utils/golang"

type Packet struct {
	Source *HID
	Dest   *HID
	Origin *HID
	MID    uint32
	PID    uint32
	M      bool
	P      uint16
	Data []byte
}

func (p *Packet) Marshal() []byte {
	ba:=NewByteArray()
	ba.Add(p.Source.Marshal())
	ba.Add(p.Dest.Marshal())
	ba.Add(p.Origin.Marshal())
	ba.AddUInt32(p.MID)
	ba.AddUInt32(p.PID)
	ba.AddBool(p.M)
	ba.AddUInt16(p.P)
	ba.AddByteArray(encrypt(p.Data))
	return ba.Data()
}

func unmarshalPacketHeader(data []byte) (*HID,*HID,*ByteArray) {
	ba:=NewByteArrayWithData(data,0)
	source:=&HID{}
	dest:=&HID{}
	source.Unmarshal(ba)
	dest.Unmarshal(ba)
	return source,dest,ba
}

func (p *Packet) UnmarshalAll(source,dest *HID,ba *ByteArray) {
	p.Source = source
	p.Dest = dest
	p.Origin = &HID{}
	p.Origin.Unmarshal(ba)
	p.MID =ba.GetUInt32()
	p.PID=ba.GetUInt32()
	p.M=ba.GetBool()
	p.P=ba.GetUInt16()
	p.Data=decrypt(ba.GetByteArray())
}