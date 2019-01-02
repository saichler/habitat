package habitat

import (
	"github.com/saichler/utils/golang"
	"github.com/sirupsen/logrus"
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
	if !in.external {
		delete(s.internal,*in.peerHID)
	} else {
		delete(s.external,in.peerHID.getHostID())
	}
	logrus.Info("Interface "+in.peerHID.String()+" was deleted")
}

func (s *Switch) addInterface(in *Interface) bool {
	s.lock.Lock()
	defer s.lock.Unlock()

	if !in.external {
		old := s.internal[*in.peerHID]
		if old!=nil {
			s.removeInterface(old)
		}
		s.internal[*in.peerHID] = in
	} else {
		old:= s.external[in.peerHID.getHostID()]
		if old!=nil {
			s.removeInterface(old)
		}
		s.external[in.peerHID.getHostID()] = in
	}

	return true
}

func (s *Switch) handlePacket(data []byte,inbox *Inbox) error {
	source,dest,ba:=unmarshalPacketHeader(data)
	if dest.IsMulticast() {
		s.handleMulticast(source,dest,data,ba,inbox)
	} else if dest.Equal(s.habitat.HID()) {
		s.handleMyPacket(source,dest,data,ba,inbox)
	} else {
		in:=s.getInterface(dest)
		in.sendData(data)
	}
	return nil
}

func (s *Switch) handleMulticast(source,dest *HID,data []byte, ba *utils.ByteArray, inbox *Inbox){
	if s.habitat.isSwitch {
		all:=s.getAllInternal()
		for k,v:=range all {
			if !k.Equal(source) {
				v.sendData(data)
			}
		}
		if source.sameMachine(s.habitat.hid) {
			all:=s.getAllExternal()
			for _,v:=range all {
				v.sendData(data)
			}
		}
	}

	s.handleMyPacket(source,dest,data,ba,inbox)
}

func (s *Switch) handleMyPacket(source,dest *HID,data []byte, ba *utils.ByteArray, inbox *Inbox){
	message := Message{}
	p:=&Packet{}
	p.UnmarshalAll(source,dest,ba)
	message.Decode(p,inbox)

	if message.Complete {
		ne:=s.getInterface(source)
		ne.statistics.mtx.Lock()
		ne.statistics.TxMessages++
		ne.statistics.mtx.Unlock()
		s.habitat.messageHandler.HandleMessage(s.habitat, &message)
	}
}

func (s *Switch) getAllInternal() map[HID]*Interface {
	s.lock.Lock()
	defer s.lock.Unlock()
	result:=make(map[HID]*Interface)
	for k,v:=range s.internal {
		if !v.isClosed {
			result[k] = v
		}
	}
	return result
}

func (s *Switch) getAllExternal() map[int32]*Interface {
	s.lock.Lock()
	defer s.lock.Unlock()
	result:=make(map[int32]*Interface)
	for k,v:=range s.external {
		if !v.isClosed {
			result[k] = v
		}
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
		if s.habitat.isSwitch {
			in = s.external[hid.getHostID()]
		} else {
			in = s.internal[*s.habitat.GetSwitchNID()]
		}
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

func (s *Switch) multicastFromSwitch(message *Message){
	faulty:=make([]*Interface,0)
	internal:=s.getAllInternal()
	for _,in:=range internal {
		err:=message.Send(in)
		if err!=nil {
			faulty = append(faulty, in)
		}
	}
	if message.Source.Hid.getHostID()==s.habitat.HID().getHostID() {
		external:=s.getAllExternal()
		for _,in:=range external {
			err:=message.Send(in)
			if err!=nil {
				faulty = append(faulty, in)
			}
		}
	}
	for _,in:=range faulty {
		s.lock.Lock()
		s.removeInterface(in)
		s.lock.Unlock()
	}
}