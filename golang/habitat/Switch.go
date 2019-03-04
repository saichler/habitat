package habitat

import (
	"errors"
	. "github.com/saichler/utils/golang"
)

type Switch struct {
	habitat  *Habitat
	internal *ConcurrentMap
	external *ConcurrentMap

}

func newSwitch(habitat *Habitat) *Switch {
	nSwitch:=&Switch{}
	nSwitch.internal = NewConcurrentMap()
	nSwitch.external = NewConcurrentMap()
	nSwitch.habitat = habitat
	return nSwitch
}

func (s *Switch) removeInterface(in *Interface) {
	if !in.external {
		s.internal.Del(*in.peerHID)
	} else {
		s.external.Del(in.peerHID.getHostID())
	}
	Info("Interface "+in.peerHID.String()+" was deleted")
}

func (s *Switch) addInterface(in *Interface) bool {
	if !in.external {
		old,_ := s.internal.Get(*in.peerHID)
		if old!=nil {
			s.removeInterface(old.(*Interface))
		}
		s.internal.Put(*in.peerHID,in)
	} else {
		old,_:=s.external.Get(in.peerHID.getHostID())
		if old!=nil {
			s.removeInterface(old.(*Interface))
		}
		s.external.Put(in.peerHID.getHostID(),in)
	}
	return true
}

func (s *Switch) sendUnreachable(source *HabitatID,priority int,data []byte) {
	in:=s.getInterface(source)
	if in==nil {
		return
	}
	p:=CreatePacket(source,UNREACH_HID,0,0,false,0,data)
	in.mailbox.outbox.Push(p.Marshal(),priority)
}

func (s *Switch) handlePacket(data []byte, mailbox *Mailbox) error {
	source,dest,m,prs,pri,ba:=unmarshalPacketHeader(data)
	if dest.Equal(UNREACH_HID) {
		if source.Equal(s.habitat.hid) {
			s.handleMyPacket(source,dest,m,prs,pri,data,ba, mailbox,true)

		}
	} else if dest.IsPublish() {
		s.handleMulticast(source,dest,m,prs,pri,data,ba, mailbox)
	} else if dest.Equal(s.habitat.HID()) {
		s.handleMyPacket(source,dest,m,prs,pri,data,ba, mailbox,false)
	} else {
		in:=s.getInterface(dest)
		if in==nil {
			s.sendUnreachable(source,pri,data)
			return errors.New("Unreachable address:"+dest.String())
		}
		in.mailbox.PushOutbox(data,pri)
	}
	return nil
}

func (s *Switch) handleMulticast(source,dest *HabitatID,m,prs bool,pri int,data []byte,ba *ByteSlice, inbox *Mailbox){
	if s.habitat.isSwitch {
		all:=s.getAllInternal()
		for k,v:=range all {
			if !k.Equal(source) {
				v.mailbox.PushOutbox(data,pri)
			}
		}
		if source.sameMachine(s.habitat.hid) {
			all:=s.getAllExternal()
			for _,v:=range all {
				v.mailbox.PushOutbox(data,pri)
			}
		}
	}
	s.handleMyPacket(source,dest,m,prs,pri,data,ba,inbox,false)
}

func (s *Switch) handleMyPacket(source,dest *HabitatID,m,prs bool,pri int,data []byte, ba *ByteSlice, inbox *Mailbox,isUnreachable bool){
	message := Message{}
	p:=&Packet{}
	p.UnmarshalAll(source,dest,m,prs,pri,ba)
	message.Decode(p,inbox,isUnreachable)

	if message.Complete {
		ne:=s.getInterface(source)
		ne.statistics.AddRxMessages()
		if !isUnreachable {
			s.habitat.messageHandler.HandleMessage(s.habitat, &message)
		} else {
			s.habitat.messageHandler.HandleUnreachable(s.habitat, &message)
		}
	}
}

func (s *Switch) getAllInternal() map[HabitatID]*Interface {
	result:=make(map[HabitatID]*Interface)
	all:=s.internal.GetMap()
	for k,v:=range all {
		key:=k.(HabitatID)
		value:=v.(*Interface)
		if !value.isClosed {
			result[key] = value
		}
	}
	return result
}

func (s *Switch) getAllExternal() map[int32]*Interface {
	result:=make(map[int32]*Interface)
	all:=s.external.GetMap()
	for k,v:=range all{
		key:=k.(int32)
		value:=v.(*Interface)
		if !value.isClosed {
			result[key] = value
		}
	}
	return result
}

func (s *Switch) getInterface(hid *HabitatID) *Interface {
	var in *Interface
	if hid.sameMachine(s.habitat.hid) {
		if s.habitat.isSwitch {
			inter,_ := s.internal.Get(*hid)
			if inter==nil {
				return nil
			}
			in = inter.(*Interface)
		} else {
			inter,_ := s.internal.Get(*s.habitat.GetSwitchNID())
			if inter==nil {
				return nil
			}
			in = inter.(*Interface)
		}
	} else {
		if s.habitat.isSwitch {
			inter,_ := s.external.Get(hid.getHostID())
			if inter==nil {
				return nil
			}
			in = inter.(*Interface)
		} else {
			inter,_ := s.internal.Get(*s.habitat.GetSwitchNID())
			if inter==nil {
				return nil
			}
			in = inter.(*Interface)
		}
	}
	return in
}

func (s *Switch) shutdown() {
	allInternal:=s.getAllInternal()
	for _,v:=range allInternal {
		v.Shutdown()
	}
	allExternal:=s.getAllExternal()
	for _,v:=range allExternal {
		v.Shutdown()
	}
}

func (s *Switch) multicastFromSwitch(message *Message){
	faulty:=make([]*Interface,0)
	internal:=s.getAllInternal()
	for _,in:=range internal {
		err:=message.Send(in)
		if err!=nil {
			faulty = append(faulty, in)
		}
	}
	if message.Source.hid.getHostID()==s.habitat.HID().getHostID() {
		external:=s.getAllExternal()
		for _,in:=range external {
			err:=message.Send(in)
			if err!=nil {
				faulty = append(faulty, in)
			}
		}
	}
	for _,in:=range faulty {
		s.removeInterface(in)
	}
}