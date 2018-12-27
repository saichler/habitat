package tests

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	log "github.com/sirupsen/logrus"
	"sync"
)

type StringMessageHandler struct {
	replyCount int
	print bool
	myx *sync.Mutex
}

type Protocol struct {
	op uint32
	data string
}

const (
	REQUEST = 1
	REPLY = 2;
)

func NewStringMessageHandler() *StringMessageHandler {
	sfh:=&StringMessageHandler{}
	sfh.print = true
	sfh.myx = &sync.Mutex{}
	return sfh
}

func getData(message *Message) *Protocol {
	ba := NewByteArrayWithData(message.Data,0)
	protocol := Protocol{}
	protocol.op = ba.GetUInt32()
	protocol.data = ba.GetString()
	return &protocol
}

func (sfh *StringMessageHandler) HandleMessage(habitat *Habitat, message *Message){
	protocol := getData(message)
	if protocol.op == REQUEST {
		if sfh.print {
			log.Info("Request: " + protocol.data+" from:"+message.Source.String())
		}
		sfh.ReplyString(protocol.data, habitat, message.Source)
	} else {
		sfh.myx.Lock()
		sfh.replyCount++
		sfh.myx.Unlock()
		if sfh.print {
			log.Info("Reply: " + protocol.data+" to:"+message.Dest.String())
		}
	}
}

func (sfh *StringMessageHandler)SendString(str string, habitat *Habitat, dest *HID){
	if sfh.print {
		log.Debug("Sending Request:" + str)
	}
	if dest==nil {
		dest = habitat.GetSwitchNID()
	}

	ba := ByteArray{}
	ba.AddUInt32(REQUEST)
	ba.AddString(str)
	message := habitat.NewMessage(habitat.GetNID(),dest,ba.Data())

	habitat.Send(message)
}

func (sfh *StringMessageHandler)ReplyString(str string, habitat *Habitat, dest *HID){
	if sfh.print {
		log.Debug("Sending Reply:"+str+" to:"+dest.String())
	}
	if dest==nil {
		dest = habitat.GetSwitchNID()
	}

	ba := ByteArray{}
	ba.AddUInt32(REPLY)
	ba.AddString(str)
	message := habitat.NewMessage(habitat.GetNID(),dest,ba.Data())

	habitat.Send(message)
}
