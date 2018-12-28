package habitat

import (
	. "github.com/saichler/utils/golang"
)

type Message struct {
	MID      uint32
	Source   *HID
	Dest     *HID
	Origin   *HID
	Data     []byte
	Complete bool
}

type MessageHandler interface {
	HandleMessage(*Habitat,*Message)
}

func (message *Message) Decode (packet *Packet, inbox *Inbox){
	message.Source = packet.Source
	message.Dest = packet.Dest

	if packet.M {
		message.Data,message.Complete=inbox.addPacket(packet)
	} else {
		message.Data = packet.Data
		message.Complete = true
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

		packet := ne.CreatePacket(message.Dest,nil,message.MID,0,true,0,ba.Data())
		ne.sendPacket(packet)

		for i:=0;i<totalParts-1;i++ {
			loc := i*MTU
			var data []byte
			if i<totalParts-2 || left==0 {
				data = message.Data[loc:loc+MTU]
			} else {
				data = message.Data[loc:loc+left]
			}

			packet := ne.CreatePacket(message.Dest,nil,message.MID,uint32(i+1),true,0,data)
			ne.sendPacket(packet)
		}
	} else {
		packet := ne.CreatePacket(message.Dest,nil,message.MID,0,false,0,message.Data)
		ne.sendPacket(packet)
	}
	return nil
}