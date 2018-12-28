package habitat

import (
	"errors"
	"github.com/saichler/security"
	log "github.com/sirupsen/logrus"
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
var ENCRYPTED   = false

type Habitat struct {
	hid          	*HID
	messageHandler  MessageHandler
	isSwitch     	bool
	nSwitch         *Switch
	netListener     net.Listener
	switchHID 		*HID
	lock *sync.Cond
	nextFrameID uint32
	running bool
}

func bind() (net.Listener,int,error){
	port := SWITCH_PORT

	log.Debug("Trying to bind to switch port " + strconv.Itoa(port) + ".");
	socket, e := net.Listen("tcp", ":"+strconv.Itoa(port))

	if e != nil {
		for ; port < MAX_PORT && e != nil; port++ {
			log.Debug("Trying to bind to port " + strconv.Itoa(port) + ".")
			s, err := net.Listen("tcp", ":"+strconv.Itoa(port))
			e = err
			socket = s
			if e==nil {
				break
			}
		}
		log.Debug("Successfuly binded to port " + strconv.Itoa(port))
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
		habitat.hid = NewHID(port)
		log.Debug("Bounded to port " + habitat.hid.String())
		habitat.isSwitch = port==SWITCH_PORT
		if !habitat.isSwitch {
			habitat.uplinkToSwitch()
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
}

func (habitat *Habitat) start() {
	go habitat.waitForlinks()
	time.Sleep(time.Second/5)
}

func (habitat *Habitat) waitForlinks() {
	if habitat.running {
		log.Info("Habitat ", habitat.hid.String(), " is waiting for links")
	}
	for ;habitat.running;{
		connection, error := habitat.netListener.Accept()
		if !habitat.running {
			break
		}
		if error != nil {
			log.Fatal("Failed to accept a new connection from socket: ", error)
			return
		}
		//add a new interface
		go habitat.addInterface(connection)
	}
	habitat.netListener.Close()
	log.Info("Habitat:"+habitat.hid.String()+" was shutdown!")
}

func (habitat *Habitat) addInterface(c net.Conn) error{
	log.Debug("connecting to: " + c.RemoteAddr().String())
	in:= newInterface(c, habitat)
	added,e:=in.handshake()
	if e!=nil {
		log.Error("Failed to add interface:",e)
	}

	if e!=nil || !added {
		return e
	}

	in.start()

	return nil
}

func (habitat *Habitat) uplinkToSwitch() {
	switchPortString := strconv.Itoa(SWITCH_PORT)
	c, e := net.Dial("tcp", "127.0.0.1:"+switchPortString)
	if e != nil {
		log.Fatal("Failed to open connection to switch: ", e)
	}
	go habitat.addInterface(c)
}

func (habitat *Habitat) Uplink(host string) {
	switchPortString := strconv.Itoa(SWITCH_PORT)
	c, e := net.Dial("tcp", host+":"+switchPortString)
	if e != nil {
		log.Fatal("Failed to open connection to host: "+host, e)
	}
	go habitat.addInterface(c)
}

func (habitat *Habitat) Send(message *Message) error {
	var e error
	if message.Dest.EqualNoCID(habitat.hid){
		habitat.messageHandler.HandleMessage(habitat,message)
	} else if message.Dest.UuidL==MULTICAST {
		if !message.Source.EqualNoCID(habitat.hid) {
			return errors.New("Multicast Message Cannot be forward!")
		}
		habitat.messageHandler.HandleMessage(habitat,message)
		ne := habitat.nSwitch.getInterface(habitat.switchHID)
		e = message.Send(ne)
		if e != nil {
			log.Error("Failed to send multicast message:", e)
		}
	} else {
		ne := habitat.nSwitch.getInterface(message.Dest)
		e = message.Send(ne)
		if e != nil {
			log.Error("Failed to send message:", e)
		}
	}
	return e
}

func (habitat *Habitat) GetSwitchNID() *HID {
	return habitat.switchHID
}

func (habitat *Habitat) GetNID() *HID {
	return habitat.hid
}

func (habitat *Habitat) NextFrameID() uint32 {
	habitat.lock.L.Lock()
	defer habitat.lock.L.Unlock()
	result:=habitat.nextFrameID
	habitat.nextFrameID++
	return result
}

func (habitat *Habitat) NewMessage(source *HID, dest *HID, data []byte) *Message {
	message := Message{}
	message.MID = habitat.NextFrameID()
	message.Source = source
	message.Dest = dest
	message.Data = data
	return &message
}

func encrypt(data []byte) ([]byte) {
	if ENCRYPTED {
		encData, err := security.Encode(data, KEY)
		if err != nil {
			log.Error("Failed to encrypt data, sending unsecure!", err)
			return data
		} else {
			return encData
		}
	}
	return data
}

func decrypt(data []byte) []byte {
	if ENCRYPTED {
		decryData,err:=security.Decode(data,KEY)
		if err!=nil {
			log.Error("Failed to decrypt data!",err)
			return data
		} else {
			return decryData
		}
	}
	return data
}