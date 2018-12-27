package habitat

import (
	"sync"
)

type Switch struct {
	node *Habitat
	internal map[HID]*Interface
	external map[int32]*Interface
	lock sync.Mutex
}

func newSwitch(node *Habitat) *Switch {
	nSwitch:=&Switch{}
	nSwitch.internal = make(map[HID]*Interface)
	nSwitch.external = make(map[int32]*Interface)
	nSwitch.node = node
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
	if p.Dest.Equal(s.node.hid) {
		message := Message{}
		message.Decode(p,inbox)
		if message.Complete {
			s.node.messageHandler.HandleMessage(s.node, &message)
		}
	} else {
		var in *Interface
		if p.Dest.sameMachine(s.node.hid) {
			in = s.internal[*p.Dest]
			if in==nil {
				panic("Cannot find:"+p.Dest.String())
			}
		} else {
			in = s.external[p.Dest.getHostID()]
		}
		in.sendPacket(p)
	}
	return nil
}

func (s *Switch) getInterface(hid *HID) *Interface {
	s.lock.Lock()
	defer s.lock.Unlock()

	var in *Interface
	if hid.sameMachine(s.node.hid) {
		if s.node.isSwitch {
			in = s.internal[*hid]
		} else {
			in = s.internal[*s.node.GetSwitchNID()]
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