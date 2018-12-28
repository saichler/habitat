package habitat

import (
	. "github.com/saichler/utils/golang"
	"strconv"
)


type SID struct {
	Hid *HID
	CID uint16
}

type Message struct {
	MID      uint32
	Source   *SID
	Dest     *SID
	Origin   *SID
	Data     []byte
	Complete bool
}

type MessageHandler interface {
	HandleMessage(*Habitat,*Message)
}

func (sid *SID) String() string {
	result:=sid.Hid.String()
	result+=strconv.Itoa(int(sid.CID))
	return result
}

func NewSID(hid *HID,cid uint16) *SID {
	sid:=&SID{}
	sid.Hid = hid
	sid.CID = cid
	return sid
}

func (message *Message) Decode (packet *Packet, inbox *Inbox){
	if packet.M {
		message.Data,message.Complete=inbox.addPacket(packet)
	} else {
		message.Data = packet.Data
		message.Complete = true
	}

	if message.Complete {
		message.Source = NewSID(packet.Source,packet.SourceSID)
		message.Dest = NewSID(packet.Dest,packet.DestSID)
		message.Origin = NewSID(packet.Origin,packet.OriginSID)
	}
}

func (message *Message) Send(ne *Interface) error {
	message.Data = encrypt(message.Data)

	if len(message.Data)> MTU {
		totalParts := len(message.Data)/MTU
		left := len(message.Data) - totalParts*MTU
		if left>0 {
			totalParts++
		}
		totalParts++

		ba := ByteArray{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(message.Data)))

		packet := ne.CreatePacket(message.Source.CID,message.Dest,message.Origin,message.MID,0,true,0,ba.Data())
		ne.sendPacket(packet)

		for i:=0;i<totalParts-1;i++ {
			loc := i*MTU
			var data []byte
			if i<totalParts-2 || left==0 {
				data = message.Data[loc:loc+MTU]
			} else {
				data = message.Data[loc:loc+left]
			}

			packet := ne.CreatePacket(message.Source.CID,message.Dest,message.Origin,message.MID,uint32(i+1),true,0,data)
			ne.sendPacket(packet)
		}
	} else {
		packet := ne.CreatePacket(message.Source.CID,message.Dest,message.Origin,message.MID,0,false,0,message.Data)
		ne.sendPacket(packet)
	}
	return nil
}