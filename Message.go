package habitat

import (
	. "github.com/saichler/utils/golang"
	"strconv"
)


type Message struct {
	MID      uint32
	Source   *ServiceID
	Dest     *ServiceID
	Origin   *ServiceID
	Type     uint16
	Data     []byte
	Complete bool
}

type MessageHandler interface {
	HandleMessage(*Habitat,*Message)
	HandleUnreachable(*Habitat,*Message)
}

func (message *Message) Decode (pkt *Packet, inbox *Mailbox,isUnreachable bool){

	packet:=pkt

	if isUnreachable {
		origSource, origDest, ba := unmarshalPacketHeader(pkt.Data)
		packet.UnmarshalAll(origSource,origDest,ba)
	}

	if packet.MultiPart {
		message.Data,message.Complete=inbox.addPacket(packet)
	} else {
		message.Data = packet.Data
		message.Complete = true
	}

	if message.Complete {
		if packet.Dest.Equal(UNREACH_HID) {
		} else {
			message.Unmarshal(packet.Source, packet.Dest)
		}
	}
}

func (message *Message) Unmarshal(source,dest *HabitatID) {
	ba:= NewByteSliceWithData(message.Data,0)
	message.Source = NewServiceID(source,ba.GetUInt16(),ba.GetString())
	message.Dest = NewServiceID(dest,ba.GetUInt16(),ba.GetString())
	message.Origin = &ServiceID{}
	message.Origin.Unmarshal(ba)
	message.Type = ba.GetUInt16()
	message.Data = ba.GetByteSlice()
}

func (message *Message) Marshal() []byte {
	ba:= NewByteSlice()
	ba.AddUInt16(message.Source.cid)
	ba.AddString(message.Source.topic)
	ba.AddUInt16(message.Dest.cid)
	ba.AddString(message.Dest.topic)
	message.Origin.Marshal(ba)
	ba.AddUInt16(message.Type)
	ba.AddByteSlice(message.Data)
	return ba.Data()
}

func (message *Message) Send(ne *Interface) error {
	ne.statistics.AddTxMessages()

	messageData:=message.Marshal()

	if len(messageData)> MTU {

		totalParts := len(messageData)/MTU
		left := len(messageData) - totalParts*MTU

		if left>0 {
			totalParts++
		}

		totalParts++

		if totalParts>1000 {
			Info("Large Message, total parts:"+strconv.Itoa(totalParts))
		}

		ba := ByteSlice{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(messageData)))

		packet := ne.CreatePacket(message.Dest,message.MID,0,true,0,ba.Data())
		err:=ne.sendPacket(packet)
		if err!=nil {
			return err
		}

		for i:=0;i<totalParts-1;i++ {
			loc := i*MTU
			var packetData []byte
			if i<totalParts-2 || left==0 {
				packetData = messageData[loc:loc+MTU]
			} else {
				packetData = messageData[loc:loc+left]
			}

			packet := ne.CreatePacket(message.Dest,message.MID,uint32(i+1),true,0, packetData)
			if i%1000==0 {
				Info("Sent "+strconv.Itoa(i)+" packets out of "+strconv.Itoa(totalParts))
			}
			err = ne.sendPacket(packet)
			if err!=nil {
				Error("Was able to send only"+strconv.Itoa(i)+" packets")
				break
			}
		}
	} else {
		packet := ne.CreatePacket(message.Dest,message.MID,0,false,0,messageData)
		ne.sendPacket(packet)
	}

	return nil
}

func (message *Message) IsPublish() bool {
	return message.Dest.IsPublish()
}