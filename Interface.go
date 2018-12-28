package habitat

import (
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"sync"
)

type Interface struct {
	habitat   *Habitat
	peerHID   *HID
	conn      net.Conn
	external  bool
	readLock  sync.Mutex
	writeLock sync.Mutex
	inbox *Inbox
}

var HANDSHAK_DATA = []byte{127,83,83,127,12,10,11}

func newInterface(conn net.Conn, habitat *Habitat) *Interface {
	in:=&Interface{}
	in.conn = conn
	in.habitat = habitat
	in.inbox = NewInbox()
	return in
}

func (in *Interface)CreatePacket(sourceSID uint16, dest,origin *SID, frameId,packetNumber uint32, multi bool, priority uint16, data []byte) *Packet {
	packet := &Packet{}
	packet.Source = in.habitat.hid
	packet.SourceSID = sourceSID
	if dest!=nil {
		packet.Dest = dest.Hid
		packet.DestSID = dest.CID
	}
	if origin!=nil {
		packet.Origin = origin.Hid
		packet.OriginSID = origin.CID
	}
	packet.MID = frameId
	packet.PID = packetNumber
	packet.M = multi
	packet.P = priority
	packet.Data = data
	return packet
}

func (in *Interface) sendData(data []byte) {
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))

	in.writeLock.Lock()
	defer in.writeLock.Unlock()

	n,e := in.conn.Write(size)
	if e!=nil || n!=len(size) {
		panic("Error! Failed to write data size bytes")
	}

	n, e = in.conn.Write(data)
	if e != nil || n != len(data) {
		panic("Error! Failed to write data bytes")
	}
}

func (in *Interface) sendPacket(p *Packet) {
	data:=p.Marshal()
	in.sendData(data)
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
}

func (in *Interface) handle() {
	for ;in.habitat.running;{
		data := in.inbox.Pop().([]byte)
		if data != nil {
			in.habitat.nSwitch.handlePacket(data,in.inbox)
		} else {
			break
		}
	}
}

func (in *Interface) start() {
	go in.read()
	go in.handle()
}

func (in *Interface) readNextPacket() error {
	in.readLock.Lock()
	defer in.readLock.Unlock()

	pSize := make([]byte, 4)
	_, e := in.conn.Read(pSize)

	if !in.habitat.running {
		return nil
	}

	if e!=nil {
		return Error("Failed to read packet size",e)
	}
	size:=int(binary.LittleEndian.Uint32(pSize))
	data := make([]byte,size)
	_, e = in.conn.Read(data)

	if !in.habitat.running {
		return nil
	}

	if e!=nil {
		return Error("Failed to read header",e)
	}

	if in.habitat.running {
		in.inbox.Push(data)
	}

	return nil
}

func (in *Interface) handshake() (bool, error) {
	log.Info("Starting handshake process for:"+in.habitat.hid.String())

	packet := in.CreatePacket(0,nil,nil,0,0,false,0,HANDSHAK_DATA)
	packet.Data = encrypt(packet.Data)

	in.sendPacket(packet)

	err:=in.readNextPacket()
	if err!=nil {
		return false,err
	}

	data:=in.inbox.Pop().([]byte)
	source,dest,ba:=unmarshalPacketHeader(data)
	p:=&Packet{}
	p.UnmarshalAll(source,dest,ba)
	p.Data = decrypt(p.Data)

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