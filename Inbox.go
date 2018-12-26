package habitat

import . "github.com/saichler/utils/golang"

type Inbox struct {
	pending map[string]map[uint32]*MultiPart
	inQueue     *Queue
}

type MultiPart struct {
	fID 	uint32
	packets []*Packet
	data    [][]byte
	totalExpectedPackets uint32
	byteLength uint32
}

func NewInbox() *Inbox {
	inbox:=&Inbox{}
	inbox.inQueue = NewQueue()
	inbox.pending = make(map[string]map[uint32]*MultiPart)
	return inbox
}

func (inbox *Inbox) Pop() interface{} {
	return inbox.inQueue.Pop()
}

func (inbox *Inbox) Push(any interface{}) {
	inbox.inQueue.Push(any)
}

func (inbox *Inbox) getMultiPart(packet *Packet) (*MultiPart,map[uint32]*MultiPart) {
	sourcePending := inbox.pending[packet.Source.FormattedString()]
	if sourcePending == nil {
		sourcePending = make(map[uint32]*MultiPart)
		inbox.pending[packet.Source.FormattedString()] = sourcePending
	}
	multiPart := sourcePending[packet.FID]
	if multiPart == nil {
		multiPart = &MultiPart{}
		multiPart.packets = make([]*Packet,0)
		sourcePending[packet.FID] = multiPart
	}
	return multiPart,sourcePending
}

func (inbox *Inbox) addPacket(packet *Packet, data []byte) ([]byte, bool) {
	multiPart,sourcePending:=inbox.getMultiPart(packet)
	multiPart.packets = append(multiPart.packets,packet)

	if multiPart.totalExpectedPackets == 0 && packet.PID == 0 {
		ba := NewByteArrayWithData(data)
		multiPart.totalExpectedPackets = ba.GetUInt32()
		multiPart.byteLength = ba.GetUInt32()
	}

	if multiPart.totalExpectedPackets>0 && len(multiPart.packets) == int(multiPart.totalExpectedPackets) {
		frameData := make([]byte,int(multiPart.byteLength))
		for i:=0;i<int(multiPart.totalExpectedPackets);i++ {
			if multiPart.packets[i].PID !=0 {
				start := int((multiPart.packets[i].PID -1)*MTU)
				end := start+len(multiPart.data[i])
				copy(frameData[start:end],multiPart.data[i][:])
			}
		}
		/* decrypt here
		key := securityutil.SecurityKey{}
		decData, err := key.Dec(packet.Data)
		if err == nil {
			frame.Data = decData
		}*/
		sourcePending[packet.FID] = nil
		return frameData,true
	}
	return nil,false
}
