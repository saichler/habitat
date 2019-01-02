package habitat

import (
	"strconv"
	"strings"
)



func FromString(str string) *HabitatID {
	nid := HabitatID{}
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

func (nid *HabitatID) sameMachine(other *HabitatID) bool {
	myip := int32(nid.UuidL >> 32)
	otherip := int32(other.UuidL >> 32)
	return myip == otherip
}

func (nid *HabitatID) getHostID() int32 {
	return int32(nid.UuidL >> 32)
}

func (nid *HabitatID) getPort() int32 {
	return int32(nid.UuidL - ((nid.UuidL >> 32) << 32))
}