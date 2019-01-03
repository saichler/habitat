package habitat

import (
	. "github.com/saichler/utils/golang"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
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
}

func (message *Message) Decode (packet *Packet, inbox *Inbox){
	if packet.M {
		message.Data,message.Complete=inbox.addPacket(packet)
	} else {
		message.Data = packet.Data
		message.Complete = true
	}

	if message.Complete {
		message.Unmarshal(packet.Source,packet.Dest)
	}
}

func (message *Message) Unmarshal(source,dest *HabitatID) {
	ba:= NewByteSliceWithData(message.Data,0)
	message.Source = NewServiceID(source,ba.GetUInt16(),ba.GetString())
	message.Dest = NewServiceID(dest,ba.GetUInt16(),ba.GetString())
	message.Origin = &ServiceID{}
	message.Origin.Unmarshal(ba)
	message.Type = ba.GetUInt16()
	message.Data = ba.GetByteArray()
}

func (message *Message) Marshal() []byte {
	ba:= NewByteSlice()
	ba.AddUInt16(message.Source.cid)
	ba.AddString(message.Source.topic)
	ba.AddUInt16(message.Dest.cid)
	ba.AddString(message.Dest.topic)
	baa:=NewByteArray()
	message.Origin.Marshal(baa)
	ba.Add(baa.Data())
	ba.AddUInt16(message.Type)
	ba.AddByteArray(message.Data)
	return ba.Data()
}

func (message *Message) Send(ne *Interface) error {
	ne.statistics.AddTxMessages()

	messageData:=message.Marshal()

	if len(messageData)> MTU {
		speedStart:=time.Now().Unix()
		bytesSent:=0

		totalParts := len(messageData)/MTU
		left := len(messageData) - totalParts*MTU

		if left>0 {
			totalParts++
		}

		totalParts++

		if totalParts>1000 {
			logrus.Info("Large Message, total parts:"+strconv.Itoa(totalParts))
		}

		ba := ByteSlice{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(messageData)))

		packet := ne.CreatePacket(message.Dest,message.MID,0,true,0,ba.Data())
		bs,err:=ne.sendPacket(packet)
		if err!=nil {
			return err
		}

		bytesSent+=bs

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
				logrus.Info("Sent "+strconv.Itoa(i)+" packets out of "+strconv.Itoa(totalParts))
			}
			bs,err = ne.sendPacket(packet)
			if err!=nil {
				logrus.Error("Was able to send only"+strconv.Itoa(i)+" packets")
				break
			}
			bytesSent+=bs
		}

		speedEnd:=time.Now().Unix()
		t:=speedEnd-speedStart
		if t>0 {
			s := float64(bytesSent)
			speed := int64(s / float64(t))
			ne.statistics.SetSpeed(speed)
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