package habitat

import (
	. "github.com/saichler/utils/golang"
)

type Frame struct {
	fID uint32
	source *HID
	dest *HID
	origin *HID
	data []byte
	complete bool
}

type FrameHandler interface {
	HandleFrame(*Habitat,*Frame)
}

func (frame *Frame) Decode (packet *Packet, inbox *Inbox){
	frame.source = packet.Source
	frame.dest = packet.Dest

	if packet.M {
		frame.data,frame.complete=inbox.addPacket(packet)
	} else {
		/* decrypt here
		key := securityutil.SecurityKey{}
		decData, err := key.Dec(packet.Data)
		if err == nil {
			frame.Data = decData
		}*/
		frame.data = packet.Data
		frame.complete = true
	}
}

func (frame *Frame) Send(ne *Interface) error {

	frameData := frame.data

	/* encrypt here
key := securityutil.SecurityKey{}
Data, err := key.Enc(packet.Data)
if err == nil {
	packet.Data = Data
}*/

	if len(frameData)> MTU {
		totalParts := len(frameData)/MTU
		left := len(frame.data) - totalParts*MTU
		if left>0 {
			totalParts++
		}
		totalParts++

		ba := ByteArray{}
		ba.AddUInt32(uint32(totalParts))
		ba.AddUInt32(uint32(len(frameData)))

		packet := ne.CreatePacket(frame.dest,nil,frame.fID,0,true,0,ba.Data())
		ne.sendPacket(packet)

		for i:=0;i<totalParts-1;i++ {
			loc := i*MTU
			var data []byte
			if i<totalParts-2 || left==0 {
				data = frameData[loc:loc+MTU]
			} else {
				data = frameData[loc:loc+left]
			}

			packet := ne.CreatePacket(frame.dest,nil,frame.fID,uint32(i+1),true,0,data)
			ne.sendPacket(packet)
		}
	} else {
		packet := ne.CreatePacket(frame.dest,nil,frame.fID,0,false,0,frame.data)
		ne.sendPacket(packet)
	}
	return nil
}

func (f *Frame) SetData(data []byte) {
	f.data = data
}

func (f *Frame) Data() []byte {
	return f.data
}

func (f *Frame) Source() *HID {
	return f.source
}