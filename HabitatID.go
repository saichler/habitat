package habitat

import (
	"bytes"
	. "github.com/saichler/utils/golang"
	"log"
	"net"
	"strconv"
	"strings"
)

const (
	PUBLISH_MARK=-9999
)

var PUBLISH_HID = newPublishHabitatID()

type HabitatID struct {
	UuidM int64
	UuidL int64
}

func NewHID(ipv4 string,port int) *HabitatID {
	newHID := &HabitatID{}
	ip:=GetIpAsInt32(ipv4)
	newHID.UuidM = 0;
	newHID.UuidL = int64(ip) << 32 + int64(port)
	return newHID
}

func NewLocalHID(port int) *HabitatID {
	return NewHID(GetLocalIpAddress(),port)
}

func newPublishHabitatID() *HabitatID {
	newHID := &HabitatID{}
	newHID.UuidM = PUBLISH_MARK
	newHID.UuidL = PUBLISH_MARK
	return newHID
}

func (hid *HabitatID) Marshal(ba *ByteArray){
	if hid!=nil {
		ba.AddInt64(hid.UuidM)
		ba.AddInt64(hid.UuidL)
	} else {
		ba.AddInt64(0)
		ba.AddInt64(0)
	}
}

func (hid *HabitatID) Unmarshal(ba *ByteArray) {
	hid.UuidM=ba.GetInt64()
	hid.UuidL=ba.GetInt64()
}

func (hid *HabitatID) String() string {
	ip := int32(hid.UuidL >> 32)
	port := int(hid.UuidL - ((hid.UuidL >> 32) << 32))
	buff:=bytes.Buffer{}
	buff.WriteString("[UuidM=")
	buff.WriteString(strconv.Itoa(int(hid.UuidM)))
	buff.WriteString(",IP=")
	buff.WriteString(GetIpAsString(ip))
	buff.WriteString(",Port=")
	buff.WriteString(strconv.Itoa(port))
	buff.WriteString("]")
	return buff.String()
}

func (hid *HabitatID) Equal(other *HabitatID) bool {
	return  hid.UuidM == other.UuidM &&
		hid.UuidL == other.UuidL
}

func (hid *HabitatID) IsPublish() bool {
	if hid.UuidL==PUBLISH_MARK {
		return true
	}
	return false
}

func GetLocalIpAddress() string {
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
				return address.String()
			}
		}
	}
	return ""
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