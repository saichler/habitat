package tests

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	log "github.com/sirupsen/logrus"
	"sync"
)

type StringFrameHandler struct {
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

func NewStringFrameHandler() *StringFrameHandler {
	sfh:=&StringFrameHandler{}
	sfh.print = true
	sfh.myx = &sync.Mutex{}
	return sfh
}

func getData(frame *Frame) *Protocol {
	ba := NewByteArrayWithData(frame.Data(),0)
	protocol := Protocol{}
	protocol.op = ba.GetUInt32()
	protocol.data = ba.GetString()
	return &protocol
}

func (sfh *StringFrameHandler) HandleFrame(habitat *Habitat, frame *Frame){
	protocol := getData(frame)
	if protocol.op == REQUEST {
		if sfh.print {
			log.Println("Request: " + protocol.data)
		}
		sfh.ReplyString(protocol.data, habitat, frame.Source())
	} else {
		sfh.myx.Lock()
		sfh.replyCount++
		sfh.myx.Unlock()
		if sfh.print {
			log.Println("Reply: " + protocol.data)
		}
	}
}

func (sfh *StringFrameHandler)SendString(str string, habitat *Habitat, dest *HID){
	if sfh.print {
		log.Debug("Sending Request:" + str)
	}
	if dest==nil {
		dest = habitat.GetSwitchNID()
	}

	ba := ByteArray{}
	ba.AddUInt32(REQUEST)
	ba.AddString(str)
	frame := habitat.NewFrame(habitat.GetNID(),dest,ba.Data())

	habitat.Send(frame)
}

func (sfh *StringFrameHandler)ReplyString(str string, habitat *Habitat, dest *HID){
	if sfh.print {
		log.Debug("Sending Reply:"+str)
	}
	if dest==nil {
		dest = habitat.GetSwitchNID()
	}

	ba := ByteArray{}
	ba.AddUInt32(REPLY)
	ba.AddString(str)
	frame := habitat.NewFrame(habitat.GetNID(),dest,ba.Data())

	habitat.Send(frame)
}
