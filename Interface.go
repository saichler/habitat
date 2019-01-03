package habitat

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
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

func (in *Interface)CreatePacket(dest *ServiceID, frameId,packetNumber uint32, multi bool, priority uint16, data []byte) *Packet {
	packet := &Packet{}
	packet.Source = in.habitat.hid
	if dest!=nil {
		packet.Dest = dest.hid
	}
	packet.MID = frameId
	packet.PID = packetNumber
	packet.M = multi
	packet.P = priority
	packet.Data = data
	return packet
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
		log.Error(msg)
		return errors.New(msg)
	}

	return nil
}

func (in *Interface) sendPacket(p *Packet) (error) {
	start:=time.Now().UnixNano()
	data:=p.Marshal()
	end:=time.Now().UnixNano()
	in.statistics.AddTxTimeSync(end-start)
	in.mailbox.PushOutbox(data)
	return nil
}

func (in *Interface) read() {
	for ;in.habitat.running;{
		err:=in.readNextPacket()
		if err!=nil {
			log.Error("Error reading from socket:", err)
			break
		}
	}
	log.Info("Interface to:"+in.peerHID.String()+" was shutdown!")
	log.Info("Statistics:")
	log.Info(in.statistics.String())
	in.isClosed = true
}

func (in *Interface) write() {
	for ;in.habitat.running;{
		data:=in.mailbox.PopOutbox()
		err:=in.sendData(data)
		if err!=nil {
			log.Error("Error Sending to socket:", err)
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
			log.Warn("Expected " + strconv.Itoa(size) + " bytes but only read 0, Sleeping a second...")
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
		in.mailbox.PushInbox(data)
	}

	return nil
}

func (in *Interface) handshake() (bool, error) {
	log.Info("Starting handshake process for:"+in.habitat.hid.String())

	packet := in.CreatePacket(nil,0,0,false,0,HANDSHAK_DATA)

	sendData:=packet.Marshal()
	in.sendData(sendData)

	err:=in.readNextPacket()
	if err!=nil {
		return false,err
	}

	data:=in.mailbox.PopInbox()

	source,dest,ba:=unmarshalPacketHeader(data)
	p:=&Packet{}
	p.UnmarshalAll(source,dest,ba)

	log.Info("handshaked "+in.habitat.hid.String()+" with nid:", p.Source.String())
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

func Error(errMsg string, e error) error {
	log.Error(e)
	return errors.New(errMsg)
}