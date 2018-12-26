package habitat

import (
	"sync"
)

type Switch struct {
	node *Habitat
	internal map[string]*Interface
	external map[int32]*Interface
	lock sync.Mutex
}

func newSwitch(node *Habitat) *Switch {
	nSwitch:=&Switch{}
	nSwitch.internal = make(map[string]*Interface)
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
		_, exist := s.internal[in.peerHID.String()]
		if exist {
			return false
		}
		s.internal[in.peerHID.String()] = in
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
		frame := Frame{}
		frame.Decode(p,inbox)
		if frame.complete {
			s.node.frameHandler.HandleFrame(s.node, &frame)
		}
	} else {
		var in *Interface
		if p.Dest.sameMachine(s.node.hid) {
			in = s.internal[p.Dest.String()]
		} else {
			in = s.external[p.Dest.getHostID()]
		}
		in.sendPacket(p)
	}
	return nil
}

func (s *Switch) getInterface(nid *HID) *Interface {
	var in *Interface
	if nid.sameMachine(s.node.hid) {
		in = s.internal[nid.String()]
	} else {
		in = s.external[nid.getHostID()]
	}
	return in
}

func (s *Switch) GetNodeSwitch(host string) *HID {
	hostID := GetIpAsInt32(host)
	return s.external[hostID].peerHID
}