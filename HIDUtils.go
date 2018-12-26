package habitat

import (
	"net"
	"log"
	"strconv"
	"strings"
)


func (nid *HID) FormattedString() string {
	ip := int32(nid.UuidL >> 32)
	port := int(nid.UuidL - ((nid.UuidL >> 32) << 32))
	return strconv.Itoa(int(nid.UuidM))+":"+GetIpAsString(ip)+":"+strconv.Itoa(port)
}

func NewHID(port int) *HID{
	newHID := &HID{}
	/*
	rand.Seed(time.Now().Unix())
	newHID.UuidMostSignificant = rand.Int63n(math.MaxInt64)
	*/
	newHID.UuidM = 0;
	var ip int32
	ip = GetIpAddress()
	newHID.UuidL = int64(ip) << 32 + int64(port)
	return newHID
}

func GetIpAddress() int32 {
	ifaces, err := net.Interfaces()
	if err!=nil {
		log.Fatal("Unable to access interfaces\n", err)
	}
	for _, _interface := range ifaces {
		intName := strings.ToLower(_interface.Name)
		if !strings.Contains(intName,"lo") &&
			!strings.Contains(intName, "br") &&
				!strings.Contains(intName, "vir") {
			intAddresses, err := _interface.Addrs()
			if err!=nil {
				log.Fatal("Unable to access interface address\n", err)
			}

			for _, address := range intAddresses {
				ipaddr := address.String()
				return GetIpAsInt32(ipaddr)
			}
		}
	}
	return 0
}

func GetIpAsString( ip int32) string {
	a := strconv.FormatInt(int64((ip>>24)&0xff), 10)
	b := strconv.FormatInt(int64((ip>>16)&0xff), 10)
	c := strconv.FormatInt(int64((ip>>8)&0xff), 10)
	d := strconv.FormatInt(int64(ip & 0xff), 10)
	return a + "." + b + "." + c + "." + d
}

func GetIpAsInt32(ipaddr string) int32 {
	var ipint int32
	arr := strings.Split(ipaddr,".")
	ipint = 0
	a,_ := strconv.Atoi(arr[0])
	b,_ := strconv.Atoi(arr[1])
	c,_ := strconv.Atoi(arr[2])
	d,_ := strconv.Atoi(strings.Split(arr[3],"/")[0])
	ipint += int32(a) << 24
	ipint += int32(b) << 16
	ipint += int32(c) << 8
	ipint += int32(d)
	return ipint
}

func FromString(str string) *HID {
	nid := HID{}
	index := strings.Index(str,":")
	mostString :=  str[0:index]
	lessString := str[index+1:len(str)]
	index1 := strings.Index(lessString,":")

	mostUUID,_ := strconv.Atoi(mostString)
	nid.UuidM = int64(mostUUID)

	ip := GetIpAsInt32(lessString[0:index1])
	port,_ := strconv.Atoi(lessString[index1+1:len(lessString)])

	nid.UuidL = int64(ip) << 32 + int64(port)
	return &nid
}

func (nid *HID) Equal (other *HID) bool {
	return  nid.UuidM == other.UuidM &&
			nid.UuidL == other.UuidL &&
			nid.Component == other.Component
}

func (nid *HID) sameMachine(other *HID) bool {
	myip := int32(nid.UuidL >> 32)
	otherip := int32(other.UuidL >> 32)
	return myip == otherip
}

func (nid *HID) getHostID() int32 {
	return int32(nid.UuidL >> 32)
}

func (nid *HID) getPort() int32 {
	return int32(nid.UuidL - ((nid.UuidL >> 32) << 32))
}