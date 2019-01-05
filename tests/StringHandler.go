package tests

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	"sync"
)

const (
	REQUEST uint16 = 1
	REPLY   uint16 = 2
)

type StringMessageHandler struct {
	replyCount int
	print bool
	myx *sync.Mutex
}

func NewStringMessageHandler() *StringMessageHandler {
	sfh:=&StringMessageHandler{}
	sfh.print = true
	sfh.myx = &sync.Mutex{}
	return sfh
}

func (sfh *StringMessageHandler) HandleMessage(habitat *Habitat, message *Message){
	str:=string(message.Data)
	if message.Type == REQUEST {
		if sfh.print {
			Info("Request: " +str+" from:"+message.Source.String())
		}
		sfh.ReplyString(str, habitat, message.Source)
	} else {
		sfh.myx.Lock()
		sfh.replyCount++
		sfh.myx.Unlock()
		if sfh.print {
			Info("Reply: " + str+" to:"+message.Dest.String())
		}
	}
}

func (sfh *StringMessageHandler)SendString(str string, habitat *Habitat, dest *ServiceID){
	if sfh.print {
		Debug("Sending Request:" + str)
	}
	if dest==nil {
		dest=NewServiceID(habitat.GetSwitchNID(),0,"")
	}
	source:=NewServiceID(habitat.HID(),0,"")
	message := habitat.NewMessage(source,dest,source,REQUEST,[]byte(str))
	habitat.Send(message)
}

func (sfh *StringMessageHandler)ReplyString(str string, habitat *Habitat, dest *ServiceID){
	if sfh.print {
		Debug("Sending Reply:"+str+" to:"+dest.String())
	}
	if dest==nil {
		dest=NewServiceID(habitat.GetSwitchNID(),0,"")
	}
	source:=NewServiceID(habitat.HID(),0,"")
	message := habitat.NewMessage(source,dest,source,REPLY,[]byte(str))

	habitat.Send(message)
}
