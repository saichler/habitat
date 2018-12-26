package tests

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	log "github.com/sirupsen/logrus"
)

type StringFrameHandler struct {
}

type Protocol struct {
	op uint32
	data string
}

const (
	REQUEST = 1
	REPLY = 2;
)

func getData(frame *Frame) *Protocol {
	ba := NewByteArrayWithData(frame.Data())
	protocol := Protocol{}
	protocol.op = ba.GetUInt32()
	protocol.data = ba.GetString()
	return &protocol
}

func (sfh *StringFrameHandler) HandleFrame(habitat *Habitat, frame *Frame){
	protocol := getData(frame)
	if protocol.op == REQUEST {
		log.Println("Request: "+protocol.data)
		sfh.ReplyString(protocol.data, habitat, frame.Source())
	} else {
		log.Println("Reply: "+protocol.data)
	}
}

func (sfh *StringFrameHandler)SendString(str string, habitat *Habitat, dest *HID){
	log.Debug("Sending Request")
	if dest==nil {
		dest = habitat.GetSwitchNID()
	}

	ba := ByteArray{}
	ba.AddUInt32(REQUEST)
	ba.AddString(str)
	frame := habitat.NewFrame(habitat.GetNID(),dest,ba.GetData())

	habitat.Send(frame)
}

func (sfh *StringFrameHandler)ReplyString(str string, habitat *Habitat, dest *HID){
	log.Debug("Sending Reply")
	if dest==nil {
		dest = habitat.GetSwitchNID()
	}

	ba := ByteArray{}
	ba.AddUInt32(REPLY)
	ba.AddString(str)
	frame := habitat.NewFrame(habitat.GetNID(),dest,ba.GetData())

	habitat.Send(frame)
}
