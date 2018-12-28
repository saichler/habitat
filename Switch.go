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

func (s *Switch) handlePacket(data []byte,inbox *Inbox) error {
	source,dest,ba:=unmarshalPacketHeader(data)
	if dest.Equal(s.habitat.hid) || dest.UuidL==MULTICAST {
		if s.habitat.isSwitch && dest.UuidL==MULTICAST {
			all:=s.getAllInternal()
			for k,v:=range all {
				if !k.Equal(source) {
					v.sendData(data)
				}
			}
		}
		message := Message{}
		p:=&Packet{}
		p.UnmarshalAll(source,dest,ba)
		message.Decode(p,inbox)
		if message.Complete {
			message.Data = decrypt(message.Data)
			s.habitat.messageHandler.HandleMessage(s.habitat, &message)
		}
	} else {
		in:=s.getInterface(dest)
		if in==nil {
			panic("cannot find:"+dest.String())
		}
		in.sendData(data)
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