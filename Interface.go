package habitat

import (
	"encoding/binary"
	"errors"
	"github.com/golang/protobuf/proto"
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

type PacketEntry struct {
	p *Packet
	header []byte
	data []byte
}

var EMPTY = make([]byte,0)

func newInterface(conn net.Conn, habitat *Habitat) *Interface {
	in:=&Interface{}
	in.conn = conn
	in.habitat = habitat
	in.inbox = NewInbox()
	return in
}

func (in *Interface)CreatePacketData(dest,origin *HID, frameId,packetNumber uint32, multi bool, priority uint32, data []byte) ([]byte,[]byte,[]byte, error) {
	packet := &Packet{}
	packet.Source = in.habitat.hid
	packet.Dest = dest
	packet.FID = frameId
	packet.PID = packetNumber
	packet.M = multi
	packet.P = priority

	header,err := proto.Marshal(packet)
	if err!=nil {
		log.Error("Failed to marshal header:", err)
		return nil,nil,nil,err
	}

	headerSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSize, uint32(len(header)))

	dataSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataSize, uint32(len(data)))

	return headerSize, header,dataSize,nil
}

func (in *Interface) sendPacket(header,data []byte) error {
	headerSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(headerSize, uint32(len(header)))

	dataSize := make([]byte, 4)
	binary.LittleEndian.PutUint32(dataSize, uint32(len(data)))

	return in.sendPacketData(headerSize,header,dataSize,data)
}

func (in *Interface) sendPacketData(headerSize,header,dataSize,data []byte) error {
	defer in.writeLock.Unlock()
	in.writeLock.Lock()
	n,e:=in.conn.Write(headerSize)
	if e!=nil || n!=len(headerSize) {
		return Error("Error! Failed to write header size bytes",e)
	}
	n,e = in.conn.Write(header)
	if e!=nil || n!=len(header) {
		return Error("Error! Failed to write header bytes:",e)
	}
	n,e = in.conn.Write(dataSize)
	if e!=nil || n!=len(dataSize) {
		return Error("Error! Failed to write data size bytes",e)
	}

	if len(data)>0 {
		n, e = in.conn.Write(data)
		if e != nil || n != len(data) {
			return Error("Error! Failed to write data bytes",e)
		}
	}
	log.Debug("Sent")
	return e
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
		pe := in.inbox.Pop().(*PacketEntry)
		if pe.data != nil {
			in.habitat.nSwitch.handlePacket(pe.p, pe.header, pe.data,in.inbox)
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

	headerSizeByte := make([]byte, 4)
	_, e := in.conn.Read(headerSizeByte)
	if e!=nil {
		return Error("Failed to read header size",e)
	}
	headerSize:=int(binary.LittleEndian.Uint32(headerSizeByte))

	header := make([]byte,headerSize)
	_, e = in.conn.Read(header)
	if e!=nil {
		return Error("Failed to read header",e)
	}

	p := &Packet{}
	e = proto.Unmarshal(header,p)
	if e!=nil {
		log.Error(e)
		return Error("Failed to unmarshel header",e)
	}

	dataSizeByte := make([]byte, 4)
	_, e = in.conn.Read(dataSizeByte)
	if e!=nil {
		return Error("Failed to read data size",e)
	}

	dataSize:=int(binary.LittleEndian.Uint32(dataSizeByte))

	if dataSize==0 {
		pe:=&PacketEntry{}
		pe.p = p
		pe.header = header
		pe.data = EMPTY
		in.inbox.Push(pe)
		return nil
	}

	data := make([]byte,dataSize)
	_, e = in.conn.Read(data)
	if e!=nil {
		return Error("Failed to read data",e)
	}

	pe:=&PacketEntry{}
	pe.p = p
	pe.header = header
	pe.data = data
	in.inbox.Push(pe)

	return nil
}

func (in *Interface) handshake() (bool, error) {
	log.Info("Starting handshake process...")

	headerSize,header,dataSize,e := in.CreatePacketData(nil,nil,0,0,false,0,EMPTY)
	if e!=nil {
		return false,e
	}

	e = in.sendPacketData(headerSize,header,dataSize,EMPTY)
	if e!=nil {
		return false,e
	}

	err:=in.readNextPacket()
	if err!=nil {
		return false,e
	}

	pe:=in.inbox.Pop().(*PacketEntry)

	log.Info("handshaked with nid:", pe.p.Source.FormattedString())
	in.peerHID = pe.p.Source
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