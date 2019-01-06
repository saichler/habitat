package habitat

import (
	. "github.com/saichler/utils/golang"
	"encoding/binary"
	"errors"
	"net"
	"strconv"
	"time"
)

type Interface struct {
	habitat    *Habitat
	peerHID    *HabitatID
	conn       net.Conn
	external   bool
	mailbox    *Mailbox
	statistics *InterfaceStatistics
	isClosed   bool
}

var HANDSHAK_DATA = []byte{127,83,83,127,12,10,11}

func newInterface(conn net.Conn, habitat *Habitat) *Interface {
	in:=&Interface{}
	in.conn = conn
	in.habitat = habitat
	in.mailbox = NewMailbox()
	in.statistics = newInterfaceStatistics()
	return in
}

func CreatePacket(source,dest *HabitatID, frameId,packetNumber uint32, multi bool, priority int, data []byte) *Packet {
	packet := &Packet{}
	packet.Source = source
	packet.Dest = dest
	packet.MID = frameId
	packet.PID = packetNumber
	packet.MultiPart = multi
	packet.Priority = priority
	packet.Data = data
	return packet
}

func (in *Interface)CreatePacket(dest *ServiceID, frameId,packetNumber uint32, multi bool, priority int, data []byte) *Packet {
	if dest!=nil {
		return CreatePacket(in.habitat.hid,dest.hid,frameId,packetNumber,multi,priority,data)
	}
	return CreatePacket(in.habitat.hid,nil,frameId,packetNumber,multi,priority,data)
}

func (in *Interface) sendData(data []byte) error {
	start:=time.Now().UnixNano()
	dataSize:=len(data)
	size:=[4]byte{}
	size[0] = byte(dataSize)
	size[1] = byte(dataSize >> 8)
	size[2] = byte(dataSize >> 16)
	size[3] = byte(dataSize >> 24)
	data = append(size[0:],data...)
	dataSize=len(data)

	in.statistics.AddTxPackets(data)

	n,e := in.conn.Write(data)

	end:=time.Now().UnixNano()

	in.statistics.AddTxTime(end-start)

	if e!=nil || n!=dataSize{
		msg:="Failed to send data: "+e.Error()
		Error(msg)
		return errors.New(msg)
	}

	return nil
}

func (in *Interface) sendPacket(p *Packet) (error) {
	start:=time.Now().UnixNano()
	data:=p.Marshal()
	end:=time.Now().UnixNano()
	in.statistics.AddTxTimeSync(end-start)
	in.mailbox.PushOutbox(data,p.Priority)
	return nil
}

func (in *Interface) read() {
	for ;in.habitat.running;{
		err:=in.readNextPacket()
		if err!=nil {
			Error("Error reading from socket:", err)
			break
		}
	}
	Info("Interface to:"+in.peerHID.String()+" was shutdown!")
	Info("Statistics:")
	Info(in.statistics.String())
	in.isClosed = true
}

func (in *Interface) write() {
	for ;in.habitat.running;{
		data:=in.mailbox.PopOutbox()
		err:=in.sendData(data)
		if err!=nil {
			Error("Error Sending to socket:", err)
			break
		}
	}
	in.isClosed = true
}

func (in *Interface) handle() {
	time.Sleep(time.Second)
	for ;in.habitat.running;{
		data := in.mailbox.PopInbox()
		if data != nil {
			in.statistics.AddRxPackets(data)
			in.habitat.nSwitch.handlePacket(data,in.mailbox)
		} else {
			break
		}
	}
}

func (in *Interface) start() {
	go in.read()
	go in.write()
	go in.handle()
}

func (in *Interface) readBytes(size int) ([]byte, error) {
	data:=make([]byte,size)
	n, e := in.conn.Read(data)

	if !in.habitat.running {
		return nil,nil
	}

	if e!=nil {
		return nil,Error("Failed to read packet size",e)
	}

	if n<size {
		if n==0 {
			Warn("Expected " + strconv.Itoa(size) + " bytes but only read 0, Sleeping a second...")
			time.Sleep(time.Second)
		}
		data=data[0:n]
		left,e:=in.readBytes(size-n)
		if e!=nil {
			return nil,Error("Failed to read packet size",e)
		}
		data = append(data,left...)
	}

	return data,nil
}


func (in *Interface) readNextPacket() error {
	//in.readLock.Lock()
	pSize,e := in.readBytes(4)
	if pSize==nil || e!=nil {
		//in.readLock.Unlock()
		return e
	}

	size:=int(binary.LittleEndian.Uint32(pSize))

	data,e := in.readBytes(size)
	if data==nil || e!=nil {
		//in.readLock.Unlock()
		return e
	}

	//in.readLock.Unlock()

	if in.habitat.running {
		in.mailbox.PushInbox(data,GetPriority(data))
	}

	return nil
}

func (in *Interface) handshake() (bool, error) {
	Info("Starting handshake process for:"+in.habitat.hid.String())

	packet := in.CreatePacket(nil,0,0,false,0,HANDSHAK_DATA)

	sendData:=packet.Marshal()
	in.sendData(sendData)

	err:=in.readNextPacket()
	if err!=nil {
		return false,err
	}

	data:=in.mailbox.PopInbox()

	source,dest,m,prs,pri,ba:=unmarshalPacketHeader(data)
	p:=&Packet{}
	p.UnmarshalAll(source,dest,m,prs,pri,ba)

	Info("handshaked "+in.habitat.hid.String()+" with nid:", p.Source.String())
	in.peerHID = p.Source
	if in.peerHID.getHostID()!=in.habitat.hid.getHostID() {
		in.external = true
	}

	if in.peerHID.getPort()==SWITCH_PORT {
		in.habitat.switchHID = in.peerHID
	}

	added:=in.habitat.nSwitch.addInterface(in)

	return added,nil
}