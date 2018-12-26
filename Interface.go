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

var EMPTY = make([]byte,0)

func newInterface(conn net.Conn, habitat *Habitat) *Interface {
	in:=&Interface{}
	in.conn = conn
	in.habitat = habitat
	in.inbox = NewInbox()
	return in
}

func (in *Interface)CreatePacket(dest,origin *HID, frameId,packetNumber uint32, multi bool, priority uint16, data []byte) *Packet {
	packet := &Packet{}
	packet.Source = in.habitat.hid
	packet.Dest = dest
	packet.FID = frameId
	packet.PID = packetNumber
	packet.M = multi
	packet.P = priority
	packet.Data = data
	return packet
}

func (in *Interface) sendPacket(p *Packet) {
	data:=p.Marshal()

	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))

	defer in.writeLock.Unlock()
	in.writeLock.Lock()

	n,e := in.conn.Write(size)
	if e!=nil || n!=len(size) {
		panic("Error! Failed to write data size bytes")
	}

	n, e = in.conn.Write(data)
	if e != nil || n != len(data) {
		panic("Error! Failed to write data bytes")
	}
}

func (in *Interface) read() {
	for ;in.habitat.running;{
		err:=in.readNextPacket()
		if err!=nil {
			log.Error("Error reading from socket:", err)
			return
		}
	}
	log.Info("Interface was shutdown!")
}

func (in *Interface) handle() {
	for ;in.habitat.running;{
		p := in.inbox.Pop().(*Packet)
		if p.Data != nil {
			in.habitat.nSwitch.handlePacket(p,in.inbox)
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
	if e!=nil {
		return Error("Failed to read packet size",e)
	}
	size:=int(binary.LittleEndian.Uint32(pSize))
	data := make([]byte,size)
	_, e = in.conn.Read(data)
	if e!=nil {
		return Error("Failed to read header",e)
	}

	p := &Packet{}
	p.Unmarshal(data)
	in.inbox.Push(p)

	return nil
}

func (in *Interface) handshake() (bool, error) {
	log.Info("Starting handshake process...")

	packet := in.CreatePacket(nil,nil,0,0,false,0,EMPTY)
	in.sendPacket(packet)

	log.Info("Starting handshake process...")

	err:=in.readNextPacket()
	if err!=nil {
		return false,err
	}

	p:=in.inbox.Pop().(*Packet)

	log.Info("handshaked with nid:", p.Source.String())
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