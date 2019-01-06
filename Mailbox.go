package habitat

import (
	. "github.com/saichler/utils/golang"
)

type Mailbox struct {
	pending *ConcurrentMap
	inbox   *PriorityQueue
	outbox  *PriorityQueue
}

type SourceMultiPart struct {
	m *ConcurrentMap
}

type MultiPart struct {
	mID                  uint32
	packets              *List
	totalExpectedPackets uint32
	byteLength           uint32
}

func NewMailbox() *Mailbox {
	mb :=&Mailbox{}
	mb.inbox = NewPriorityQueue()
	mb.outbox = NewPriorityQueue()
	mb.pending = NewConcurrentMap()
	return mb
}

func newSourceMultipart() *SourceMultiPart {
	smp:=&SourceMultiPart{}
	smp.m = NewConcurrentMap()
	return smp
}

func (mailbox *Mailbox) SetName(name string) {
	mailbox.inbox.SetName(name)
	mailbox.outbox.SetName(name)
}

func (mailbox *Mailbox) PopInbox() []byte {
	next:=mailbox.inbox.Pop()
	if next!=nil {
		return next.([]byte)
	}
	return nil
}

func (mailbox *Mailbox) PopOutbox() []byte {
	next:=mailbox.outbox.Pop()
	if next!=nil {
		return next.([]byte)
	}
	return nil
}

func (mailbox *Mailbox) PushInbox(pData []byte,priority int) {
	mailbox.inbox.Push(pData,priority)
}

func (mailbox *Mailbox) PushOutbox(pData []byte,priority int) {
	mailbox.outbox.Push(pData,priority)
}

func (smp *SourceMultiPart) newMultiPart(fid uint32) *MultiPart {
	multiPart := &MultiPart{}
	multiPart.packets = NewList()
	smp.m.Put(fid,multiPart)
	return multiPart
}

func (smp *SourceMultiPart) getMultiPart(mid uint32) *MultiPart {
	var multiPart *MultiPart
	exist,ok := smp.m.Get(mid)
	if !ok {
		multiPart = smp.newMultiPart(mid)
	} else {
		multiPart = exist.(*MultiPart)
	}
	return multiPart
}

func (smp *SourceMultiPart) delMultiPart(mid uint32) {
	smp.m.Del(mid)
}

func (mailbox *Mailbox) getMultiPart(packet *Packet) (*MultiPart,*SourceMultiPart) {
	hk:=packet.Source
	var sourceMultiParts *SourceMultiPart
	existing,ok:= mailbox.pending.Get(*hk)
	if !ok {
		sourceMultiParts=newSourceMultipart()
		mailbox.pending.Put(*hk,sourceMultiParts)
	} else {
		sourceMultiParts=existing.(*SourceMultiPart)
	}
	multiPart:= sourceMultiParts.getMultiPart(packet.MID)
	return multiPart,sourceMultiParts
}

func (mailbox *Mailbox) addPacket(packet *Packet) ([]byte, bool) {
	mp,smp:= mailbox.getMultiPart(packet)
	mp.packets.Add(packet)
	if mp.totalExpectedPackets == 0 && packet.PID == 0 {
		ba := NewByteSliceWithData(packet.Data,0)
		mp.totalExpectedPackets = ba.GetUInt32()
		mp.byteLength = ba.GetUInt32()
	}

	isComplete:=false
	if mp.totalExpectedPackets>0 && mp.packets.Size() == int(mp.totalExpectedPackets) {
		isComplete = true
	}

	if isComplete {
		messageData := make([]byte,int(mp.byteLength))
		for i:=0;i<int(mp.totalExpectedPackets);i++ {
			qp:=mp.packets.Get(i).(*Packet)
			if qp.PID !=0 {
				start := int((qp.PID -1)*uint32(MTU))
				end := start+len(qp.Data)
				copy(messageData[start:end],qp.Data[:])
			}
		}
		smp.delMultiPart(packet.MID)
		return messageData,isComplete
	}
	return nil,isComplete
}

func (mailbox *Mailbox) Shutdown() {
	mailbox.inbox.Shutdown()
	mailbox.outbox.Shutdown()
}
