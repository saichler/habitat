package habitat

import (
	"errors"
	. "github.com/saichler/security"
	."github.com/saichler/utils/golang"
	"net"
	"strconv"
	"sync"
	"time"
)

const (
	SWITCH_PORT = 52000
	MAX_PORT    = 54000
)

var MTU 		= 512
var KEY         = "bNhDNirkahDbiJJirSfaNNEXDprtwQoK"
var ENCRYPTED   = true

type Habitat struct {
	hid            *HabitatID
	messageHandler MessageHandler
	isSwitch       bool
	nSwitch        *Switch
	netListener    net.Listener
	switchHID      *HabitatID
	lock           *sync.Cond
	nextMessageID  uint32
	running        bool
}

func bind() (net.Listener,int,error){
	port := SWITCH_PORT

	Debug("Trying to bind to switch port " + strconv.Itoa(port) + ".");
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(port))

	if e != nil {
		for ; port < MAX_PORT && e != nil; port++ {
			Debug("Trying to bind to port " + strconv.Itoa(port) + ".")
			s, err := net.Listen("tcp", ":"+strconv.Itoa(port))
			e = err
			socket = s
			if e==nil {
				break
			}
		}
		Debug("Successfuly binded to port " + strconv.Itoa(port))
	}

	if port >= MAX_PORT {
		return nil,-1,errors.New("Failed to find an available port to bind to")
	}

	return socket,port,nil
}

func NewHabitat(handler MessageHandler) (*Habitat, error) {
	habitat :=&Habitat{}
	habitat.nSwitch = newSwitch(habitat)
	habitat.messageHandler = handler

	socket,port,e:=bind()

	if e != nil {
		return nil,e
	} else {
		habitat.hid = NewLocalHID(port)
		Debug("Bounded to port " + habitat.hid.String())
		habitat.isSwitch = port==SWITCH_PORT
		if !habitat.isSwitch {
			e:=habitat.uplinkToSwitch()
			for ;e!=nil; {
				time.Sleep(time.Second*5)
				e=habitat.uplinkToSwitch()
			}
		}
	}
	habitat.netListener = socket
	habitat.lock = sync.NewCond(&sync.Mutex{})
	habitat.running = true
	habitat.start()
	return habitat, nil
}

func (habitat *Habitat) Shutdown() {
	habitat.running = false
	net.Dial("tcp", "127.0.0.1:"+strconv.Itoa(int(habitat.hid.getPort())))
	habitat.nSwitch.shutdown()
	habitat.lock.L.Lock()
	habitat.lock.Broadcast()
	habitat.lock.L.Unlock()
}

func (habitat *Habitat) start() {
	go habitat.waitForlinks()
	time.Sleep(time.Second/5)
}

func (habitat *Habitat) waitForlinks() {
	if habitat.running {
		Info("Habitat ", habitat.hid.String(), " is waiting for links")
	}
	for ;habitat.running;{
		connection, error := habitat.netListener.Accept()
		if !habitat.running {
			break
		}
		if error != nil {
			Fatal("Failed to accept a new connection from socket: ", error)
			return
		}
		//add a new interface
		go habitat.addInterface(connection)
	}
	habitat.netListener.Close()
	Info("Habitat:"+habitat.hid.String()+" was shutdown!")
}

func (habitat *Habitat) addInterface(c net.Conn) (*HabitatID,error) {
	Debug("connecting to: " + c.RemoteAddr().String())
	in:= newInterface(c, habitat)
	added,e:=in.handshake()
	if e!=nil {
		Error("Failed to add interface:",e)
	}

	if e!=nil || !added {
		return nil,e
	}

	in.start()

	return in.peerHID,nil
}

func (habitat *Habitat) uplinkToSwitch() error {
	switchPortString := strconv.Itoa(SWITCH_PORT)
	c, e := net.Dial("tcp", "127.0.0.1:"+switchPortString)
	if e != nil {
		Error("Failed to open connection to switch: ", e)
		return e
	}
	go habitat.addInterface(c)
	return e
}

func (habitat *Habitat) Uplink(host string) *HabitatID {
	switchPortString := strconv.Itoa(SWITCH_PORT)
	c, e := net.Dial("tcp", host+":"+switchPortString)
	if e != nil {
		Error("Failed to open connection to host: "+host, e)
	}
	hid,err:=habitat.addInterface(c)
	if err!=nil {
		return nil
	}
	return hid
}

func (habitat *Habitat) waitForUplinkToSwitch() *Interface {
	habitat.lock.L.Lock()
	defer habitat.lock.L.Unlock()
	ne := habitat.nSwitch.getInterface(habitat.switchHID)
	if ne==nil || ne.isClosed {
		Error("Uplink to switch is closed, trying to open a new one.")
		e := habitat.uplinkToSwitch()
		for ; e != nil; {
			time.Sleep(time.Second * 5)
			e = habitat.uplinkToSwitch()
		}
	}
	time.Sleep(time.Second)
	ne = habitat.nSwitch.getInterface(habitat.switchHID)
	return ne
}

func (habitat *Habitat) Send(message *Message) error {
	var e error
	if message.Dest.hid.Equal(habitat.hid){
		habitat.messageHandler.HandleMessage(habitat,message)
	} else if message.IsPublish() {
		if !message.Source.hid.Equal(habitat.hid) {
			return errors.New("Multicast Message Cannot be forward!")
		}
		habitat.messageHandler.HandleMessage(habitat,message)
		if habitat.isSwitch {
			habitat.nSwitch.multicastFromSwitch(message)
		} else {
			ne := habitat.nSwitch.getInterface(habitat.switchHID)
			if ne==nil || ne.isClosed {
				ne = habitat.waitForUplinkToSwitch()
			}
			e = message.Send(ne)
			if e != nil {
				Error("Failed to send multicast message:", e)
			}
		}
	} else {
		ne := habitat.nSwitch.getInterface(message.Dest.hid)
		if ne==nil {
			Error("Unknown Destination:"+message.Dest.String())
			habitat.messageHandler.HandleUnreachable(habitat,message)
			return errors.New("Unknown Destination:"+message.Dest.String())
		}
		e = message.Send(ne)
		if e != nil {
			Error("Failed to send message:", e)
		}
	}
	return e
}

func (habitat *Habitat) GetSwitchNID() *HabitatID {
	return habitat.switchHID
}

func (habitat *Habitat) HID() *HabitatID {
	return habitat.hid
}

func (habitat *Habitat) ServiceID() *ServiceID {
	return NewServiceID(habitat.hid,0,"Habitat")
}

func (habitat *Habitat) nextMID() uint32 {
	habitat.lock.L.Lock()
	defer habitat.lock.L.Unlock()
	result:=habitat.nextMessageID
	habitat.nextMessageID++
	return result
}

func (habitat *Habitat) NewMessage(source, dest, origin *ServiceID,ptype uint16, data []byte) *Message {
	message := Message{}
	message.MID = habitat.nextMID()
	message.Source = source
	message.Dest = dest
	message.Origin = origin
	message.Data = data
	message.Type = ptype
	return &message
}

func (habitat *Habitat) WaitForShutdown(){
	habitat.lock.L.Lock()
	habitat.lock.Wait()
	habitat.lock.L.Unlock()
}

func (habitat *Habitat) Running() bool {
	return habitat.running
}

func encrypt(data []byte) ([]byte) {
	if ENCRYPTED {
		encData, err := Encode(data, KEY)
		if err != nil {
			Error("Failed to encrypt data, sending unsecure!", err)
			return data
		} else {
			return encData
		}
	}
	return data
}

func decrypt(data []byte) []byte {
	if ENCRYPTED {
		decryData, err := Decode(data, KEY)
		if err != nil {
			panic("Failed to decrypt data!")
			return data
		} else {
			return decryData
		}
	}
	return data
}