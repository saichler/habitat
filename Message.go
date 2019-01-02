package habitat

import (
	. "github.com/saichler/utils/golang"
	"github.com/sirupsen/logrus"
	"strconv"
	"time"
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
	Type     uint16
	Data     []byte
	Complete bool
}

type MessageHandler interface {
	HandleMessage(*Habitat,*Message)
}

func (sid *SID) String() string {
	result:=sid.Hid.String()
	result+=":"+strconv.Itoa(int(sid.CID))
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
		message.Unmarshal(packet)
	}
}

func (message *Message) Unmarshal(onePacket *Packet) {
	ba:=NewByteArrayWithData(message.Data,0)
	message.Source = NewSID(onePacket.Source,ba.GetUInt16())
	message.Dest = NewSID(onePacket.Dest,ba.GetUInt16())
	message.Origin = NewSID(onePacket.Origin,ba.GetUInt16())
	message.Type = ba.GetUInt16()
	message.Data = ba.GetByteArray()
}

func (message *Message) Marshal() []byte {
	ba:=NewByteArray()
	ba.AddUInt16(message.Source.CID)
	ba.AddUInt16(message.Dest.CID)
	ba.AddUInt16(message.Origin.CID)
	ba.AddUInt16(message.Type)
	ba.AddByteArray(message.Data)
	return ba.Data()
}

func (message *Message) Send(ne *Interface) error {
	ne.statistics.mtx.Lock()
	ne.statistics.TxMessages++
	ne.statistics.mtx.Unlock()

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

		ba := ByteArray{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(messageData)))

		packet := ne.CreatePacket(message.Dest,message.Origin,message.MID,0,true,0,ba.Data())
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

			packet := ne.CreatePacket(message.Dest,message.Origin,message.MID,uint32(i+1),true,0, packetData)
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
			if ne.statistics.AvgSpeed==0 || speed<ne.statistics.AvgSpeed{
				ne.statistics.AvgSpeed = speed
			}
		}
	} else {
		packet := ne.CreatePacket(message.Dest,message.Origin,message.MID,0,false,0,messageData)
		ne.sendPacket(packet)
	}

	return nil
}

func (message *Message) IsMulticast() bool {
	return message.Dest.Hid.IsMulticast()
}