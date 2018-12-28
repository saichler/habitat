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
	Data   []byte
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
	ba.AddByteArray(p.Data)
	return ba.Data()
}

func (p *Packet) Unmarshal(data []byte) {
	p.Source = &HID{}
	p.Dest = &HID{}
	p.Origin = &HID{}
	ba:=NewByteArrayWithData(data,0)
	p.Source.Unmarshal(ba)
	p.Dest.Unmarshal(ba)
	p.Origin.Unmarshal(ba)
	p.MID =ba.GetUInt32()
	p.PID=ba.GetUInt32()
	p.M=ba.GetBool()
	p.P=ba.GetUInt16()
	p.Data=ba.GetByteArray()
}

func unmarshalToPacket(data []byte) *Packet {
	p:=&Packet{}
	p.Unmarshal(data)
	return p
}