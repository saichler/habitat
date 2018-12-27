package habitat

import (
	"sync"
)

type Switch struct {
	habitat  *Habitat
	internal map[HID]*Interface
	external map[int32]*Interface
	lock     sync.Mutex
}

func newSwitch(habitat *Habitat) *Switch {
	nSwitch:=&Switch{}
	nSwitch.internal = make(map[HID]*Interface)
	nSwitch.external = make(map[int32]*Interface)
	nSwitch.habitat = habitat
	return nSwitch
}

func (s *Switch) removeInterface(in *Interface) {

}

func (s *Switch) addInterface(in *Interface) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !in.external {
		_, exist := s.internal[*in.peerHID]
		if exist {
			return false
		}
		s.internal[*in.peerHID] = in
	} else {
		_, exist := s.external[in.peerHID.getHostID()]
		if exist {
			return false
		}
		s.external[in.peerHID.getHostID()] = in
	}

	return true
}

func (s *Switch) handlePacket(p *Packet,inbox *Inbox) error {
	if p.Dest.EqualNoCID(s.habitat.hid) || p.Dest.UuidL==MULTICAST {
		if s.habitat.isSwitch && p.Dest.UuidL==MULTICAST {
			all:=s.getAllInternal()
			for k,v:=range all {
				if !k.EqualNoCID(p.Source) {
					v.sendPacket(p)
				}
			}
		}
		message := Message{}
		message.Decode(p,inbox)
		if message.Complete {
			s.habitat.messageHandler.HandleMessage(s.habitat, &message)
		}
	} else {
		in:=s.getInterface(p.Dest)
		/*
		var in *Interface
		if p.Dest.sameMachine(s.habitat.hid) {
			in = s.internal[*p.Dest]
			if in==nil {
				panic("Cannot find Internal:"+p.Dest.String())
			}
		} else {
			in = s.external[p.Dest.getHostID()]
			if in==nil {
				panic("Cannot find External:"+p.Dest.String())
			}
		}*/
		in.sendPacket(p)
	}
	return nil
}

func (s *Switch) getAllInternal() map[HID]*Interface {
	s.lock.Lock()
	defer s.lock.Unlock()
	result:=make(map[HID]*Interface)
	for k,v:=range s.internal {
		result[k]=v
	}
	return result
}

func (s *Switch) getInterface(hid *HID) *Interface {
	s.lock.Lock()
	defer s.lock.Unlock()

	var in *Interface
	if hid.sameMachine(s.habitat.hid) {
		if s.habitat.isSwitch {
			in = s.internal[*hid]
		} else {
			in = s.internal[*s.habitat.GetSwitchNID()]
		}
	} else {
		in = s.external[hid.getHostID()]
	}
	return in
}

func (s *Switch) GetNodeSwitch(host string) *HID {
	hostID := GetIpAsInt32(host)
	return s.external[hostID].peerHID
}

func (s *Switch) shutdown() {
	all:=s.getAllInternal()
	for _,v:=range all {
		v.conn.Close()
	}
}