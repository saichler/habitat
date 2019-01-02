package habitat

import (
	"bytes"
	"encoding/binary"
	"errors"
	log "github.com/sirupsen/logrus"
	"net"
	"strconv"
	"sync"
	"time"
)

type Interface struct {
	habitat   *Habitat
	peerHID   *HabitatID
	conn      net.Conn
	external  bool
	readLock  sync.Mutex
	writeLock sync.Mutex
	inbox *Inbox
	statistics *InterfaceStatistics
	isClosed bool
}

type InterfaceStatistics struct {
	TxMessages int64
	RxMessages int64
	TxPackets  int64
	RxPackets  int64
	TxBytes    int64
	RxBytes    int64
	AvgSpeed   int64
	mtx *sync.Mutex
}

var HANDSHAK_DATA = []byte{127,83,83,127,12,10,11}

func (ist *InterfaceStatistics) String() string {
	buff:=&bytes.Buffer{}
	buff.WriteString("Rx Messages:"+strconv.Itoa(int(ist.RxMessages)))
	buff.WriteString(" Tx Messages:"+strconv.Itoa(int(ist.TxMessages)))
	buff.WriteString(" Rx Packets:"+strconv.Itoa(int(ist.RxPackets)))
	buff.WriteString(" Tx Packets:"+strconv.Itoa(int(ist.TxPackets)))
	buff.WriteString(" Rx Bytes:"+strconv.Itoa(int(ist.RxBytes)))
	buff.WriteString(" Tx Bytes:"+strconv.Itoa(int(ist.TxBytes)))
	buff.WriteString(" Avg Speed:"+ist.getSpeed())
	return buff.String()
}

func (ist *InterfaceStatistics) getSpeed() string {
	speed:=float64(ist.AvgSpeed)
	if int64(speed)/1024==0 {
		return strconv.Itoa(int(speed))+" Bytes/Sec"
	}
	speed=speed/1024
	if int64(speed)/1024==0 {
		return strconv.Itoa(int(speed))+" Kilo Bytes/Sec"
	}
	speed=speed/1024
	s:=strconv.FormatFloat(speed, 'f', 2, 64)
	return s+" Mega Bytes/Sec"
}

func newInterface(conn net.Conn, habitat *Habitat) *Interface {
	in:=&Interface{}
	in.conn = conn
	in.habitat = habitat
	in.inbox = NewInbox()
	in.statistics = &InterfaceStatistics{}
	in.statistics.mtx = &sync.Mutex{}
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
	size := make([]byte, 4)
	binary.LittleEndian.PutUint32(size, uint32(len(data)))

	in.writeLock.Lock()
	defer in.writeLock.Unlock()

	size = append(size,data...)

	n,e := in.conn.Write(size)

	if e!=nil || n!=len(size){
		log.Error("Failed to send data:",e)
		panic("")
		return e
	}

	if n!=len(size){
		log.Info("Did not send all, send only "+strconv.Itoa(n)+ "out of "+strconv.Itoa(len(size)))
		time.Sleep(time.Millisecond*100)
	}
	return nil
}

func (in *Interface) sendPacket(p *Packet) (int,error) {
	data:=p.Marshal()
	in.statistics.mtx.Lock()
	in.statistics.TxPackets++
	size:=len(data)+4
	in.statistics.TxBytes+=int64(size)
	in.statistics.mtx.Unlock()
	return size,in.sendData(data)
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

func (in *Interface) handle() {
	time.Sleep(time.Second)
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

	in.readLock.Lock()
	pSize,e := in.readBytes(4)
	if pSize==nil || e!=nil {
		return e
	}

	size:=int(binary.LittleEndian.Uint32(pSize))

	data,e := in.readBytes(size)
	if data==nil || e!=nil {
		return e
	}
	in.readLock.Unlock()

	if in.habitat.running {
		in.statistics.mtx.Lock()
		in.statistics.RxPackets++
		in.statistics.RxBytes+=int64(len(data))
		in.statistics.mtx.Unlock()
		in.inbox.Push(data)
	}

	return nil
}

func (in *Interface) handshake() (bool, error) {
	log.Info("Starting handshake process for:"+in.habitat.hid.String())

	packet := in.CreatePacket(nil,0,0,false,0,HANDSHAK_DATA)

	_,err:=in.sendPacket(packet)
	if err!=nil {
		return false,err
	}

	err=in.readNextPacket()
	if err!=nil {
		return false,err
	}

	data:=in.inbox.Pop().([]byte)
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