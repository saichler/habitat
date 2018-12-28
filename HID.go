package habitat

import (
	. "github.com/saichler/utils/golang"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	MULTICAST=-9999
)

type HID struct {
	UuidM int64
	UuidL int64
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

func NewMulticastHID(multicastGroup int16) *HID {
	newHID := &HID{}
	newHID.UuidM = int64(multicastGroup)
	newHID.UuidL = MULTICAST
	return newHID
}

func (hid *HID) Marshal() []byte {
	ba := NewByteArray()
	if hid!=nil {
		ba.AddInt64(hid.UuidM)
		ba.AddInt64(hid.UuidL)
	} else {
		ba.AddInt64(0)
		ba.AddInt64(0)
	}
	return ba.Data()
}

func (hid *HID) Unmarshal(ba *ByteArray) {
	hid.UuidM=ba.GetInt64()
	hid.UuidL=ba.GetInt64()
}

func (hid *HID) String() string {
	ip := int32(hid.UuidL >> 32)
	port := int(hid.UuidL - ((hid.UuidL >> 32) << 32))
	return strconv.Itoa(int(hid.UuidM))+":"+GetIpAsString(ip)+":"+strconv.Itoa(port)
}

func (nid *HID) Equal(other *HID) bool {
	return  nid.UuidM == other.UuidM &&
		nid.UuidL == other.UuidL
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