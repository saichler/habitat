package habitat

import (
	"encoding/binary"
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
		message.Type = binary.LittleEndian.Uint16(message.Data[:2])
		message.Data = message.Data[2:]
		message.Source = NewSID(packet.Source,packet.SourceSID)
		message.Dest = NewSID(packet.Dest,packet.DestSID)
		message.Origin = NewSID(packet.Origin,packet.OriginSID)
	}
}

func (message *Message) Send(ne *Interface) error {
	ne.statistics.mtx.Lock()
	ne.statistics.TxMessages++
	ne.statistics.mtx.Unlock()

	mt := make([]byte, 2)
	binary.LittleEndian.PutUint16(mt, message.Type)
	message.Data = append(mt,message.Data...)

	if len(message.Data)> MTU {
		speedStart:=time.Now().Unix()
		bytesSent:=0

		totalParts := len(message.Data)/MTU
		left := len(message.Data) - totalParts*MTU
		if left>0 {
			totalParts++
		}
		totalParts++

		if totalParts>1000 {
			logrus.Info("Large Message, total parts:"+strconv.Itoa(totalParts))
		}

		ba := ByteArray{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(message.Data)))

		packet := ne.CreatePacket(message.Source.CID,message.Dest,message.Origin,message.MID,0,true,0,ba.Data())
		bs,err:=ne.sendPacket(packet)
		if err!=nil {
			return err
		}

		bytesSent+=bs

		for i:=0;i<totalParts-1;i++ {
			loc := i*MTU
			var data []byte
			if i<totalParts-2 || left==0 {
				data = message.Data[loc:loc+MTU]
			} else {
				data = message.Data[loc:loc+left]
			}

			packet := ne.CreatePacket(message.Source.CID,message.Dest,message.Origin,message.MID,uint32(i+1),true,0,data)
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
		packet := ne.CreatePacket(message.Source.CID,message.Dest,message.Origin,message.MID,0,false,0,message.Data)
		ne.sendPacket(packet)
	}

	return nil
}

func (message *Message) IsMulticast() bool {
	return message.Dest.Hid.IsMulticast()
}