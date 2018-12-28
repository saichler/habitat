package habitat

import (
	. "github.com/saichler/utils/golang"
	"sync"
)

type Inbox struct {
	pending map[HID]*SourceMultiPart
	mtx *sync.Mutex
	inQueue     *Queue
}

type SourceMultiPart struct {
	m map[uint32]*MultiPart
	mtx *sync.Mutex
}

type MultiPart struct {
	mID                  uint32
	packets              []*Packet
	totalExpectedPackets uint32
	byteLength           uint32
	mtx                  *sync.Mutex
}

func NewInbox() *Inbox {
	inbox:=&Inbox{}
	inbox.inQueue = NewQueue()
	inbox.pending = make(map[HID]*SourceMultiPart)
	inbox.mtx = &sync.Mutex{}
	return inbox
}

func newSourceMultipart() *SourceMultiPart {
	smp:=&SourceMultiPart{}
	smp.mtx = &sync.Mutex{}
	smp.m = make(map[uint32]*MultiPart)
	return smp
}

func (inbox *Inbox) Pop() interface{} {
	return inbox.inQueue.Pop()
}

func (inbox *Inbox) Push(any interface{}) {
	inbox.inQueue.Push(any)
}

func (smp *SourceMultiPart) newMultiPart(fid uint32) *MultiPart {
	multiPart := &MultiPart{}
	multiPart.packets = make([]*Packet,0)
	multiPart.mtx = &sync.Mutex{}
	smp.m[fid] = multiPart
	return multiPart
}

func (smp *SourceMultiPart) getMultiPart(mid uint32) *MultiPart {
	smp.mtx.Lock()
	defer smp.mtx.Unlock()
	multiPart := smp.m[mid]
	if multiPart == nil {
		multiPart = smp.newMultiPart(mid)
	}
	return multiPart
}

func (smp *SourceMultiPart) delMultiPart(mid uint32) {
	smp.mtx.Lock()
	defer smp.mtx.Unlock()
	delete(smp.m, mid)
}

func (inbox *Inbox) getMultiPart(packet *Packet) (*MultiPart,*SourceMultiPart) {
	hk:=packet.Source
	inbox.mtx.Lock()
	sourceMultiParts:=inbox.pending[*hk]
	if sourceMultiParts==nil {
		sourceMultiParts=newSourceMultipart()
		inbox.pending[*hk] = sourceMultiParts
	}
	inbox.mtx.Unlock()
	multiPart:= sourceMultiParts.getMultiPart(packet.MID)
	return multiPart,sourceMultiParts
}

func (inbox *Inbox) addPacket(packet *Packet) ([]byte, bool) {
	mp,smp:=inbox.getMultiPart(packet)
	mp.mtx.Lock()
	mp.packets = append(mp.packets,packet)
	if mp.totalExpectedPackets == 0 && packet.PID == 0 {
		ba := NewByteArrayWithData(packet.Data,0)
		mp.totalExpectedPackets = ba.GetUInt32()
		mp.byteLength = ba.GetUInt32()
	}

	isComplete:=false
	if mp.totalExpectedPackets>0 && len(mp.packets) == int(mp.totalExpectedPackets) {
		isComplete = true
	}
	mp.mtx.Unlock()

	if isComplete {
		messageData := make([]byte,int(mp.byteLength))
		for i:=0;i<int(mp.totalExpectedPackets);i++ {
			if mp.packets[i].PID !=0 {
				start := int((mp.packets[i].PID -1)*uint32(MTU))
				end := start+len(mp.packets[i].Data)
				copy(messageData[start:end],mp.packets[i].Data[:])
			}
		}
		/* decrypt here
		key := securityutil.SecurityKey{}
		decData, err := key.Dec(packet.Data)
		if err == nil {
			frame.Data = decData
		}*/
		smp.delMultiPart(packet.MID)
		return messageData,isComplete
	}
	return nil,isComplete
}
