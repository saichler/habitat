package tests

import (
	. "github.com/saichler/habitat"
	log "github.com/sirupsen/logrus"
	"sync"
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
	if message.Type == ProtocolType_Request {
		if sfh.print {
			log.Info("Request: " +str+" from:"+message.Source.String())
		}
		sfh.ReplyString(str, habitat, message.Source)
	} else {
		sfh.myx.Lock()
		sfh.replyCount++
		sfh.myx.Unlock()
		if sfh.print {
			log.Info("Reply: " + str+" to:"+message.Dest.String())
		}
	}
}

func (sfh *StringMessageHandler)SendString(str string, habitat *Habitat, dest *SID){
	if sfh.print {
		log.Debug("Sending Request:" + str)
	}
	if dest==nil {
		dest=NewSID(habitat.GetSwitchNID(),0)
	}
	source:=habitat.SID()
	message := habitat.NewMessage(source,dest,nil,ProtocolType_Request,[]byte(str))
	habitat.Send(message)
}

func (sfh *StringMessageHandler)ReplyString(str string, habitat *Habitat, dest *SID){
	if sfh.print {
		log.Debug("Sending Reply:"+str+" to:"+dest.String())
	}
	if dest==nil {
		dest=NewSID(habitat.GetSwitchNID(),0)
	}
	source:=habitat.SID()
	message := habitat.NewMessage(source,dest,nil,ProtocolType_Reply,[]byte(str))

	habitat.Send(message)
}
