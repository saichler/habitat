package tests

import (
	. "github.com/saichler/habitat"
	. "github.com/saichler/utils/golang"
	"os"
	"strconv"
	"testing"
	"time"
)

func setup(){
	//SetLevel(DebugLevel)
}
func tearDown(){}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}

func TestHabitat(t *testing.T) {
	MTU = 512
	ENCRYPTED=false
	h:= NewStringMessageHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	Info("Node1:",n1.ServiceID().String()," Node2:",n2.ServiceID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.ServiceID())
	h.SendString("Hello World",n2,n1.ServiceID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
	n1.Shutdown()
	n2.Shutdown()
	time.Sleep(time.Second*2)
}

func TestSwitch(t *testing.T) {
	MTU = 512
	ENCRYPTED=true
	ENCRYPTED=false
	h:= NewStringMessageHandler()

	s,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}

	n1,e:=NewHabitat(h)
	n2,e:=NewHabitat(h)

	Info("Node1:",n1.ServiceID().String()," Node2:",n2.ServiceID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.ServiceID())
	h.SendString("Hello World",n2,n1.ServiceID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
	n1.Shutdown()
	n2.Shutdown()
	s.Shutdown()
	time.Sleep(time.Second*2)
}

func TestMultiPart(t *testing.T) {
	MTU = 4
	ENCRYPTED=false
	h:= NewStringMessageHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	Info("Node1:",n1.ServiceID().String()," Node2:",n2.ServiceID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.ServiceID())
	h.SendString("Hello World",n2,n1.ServiceID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
	n1.Shutdown()
	n2.Shutdown()
	time.Sleep(time.Second*2)
}

func TestMessageScale(t *testing.T) {
	MTU = 512
	ENCRYPTED=false
	numOfMessages:=10000
	numOfHabitats:=3

	h:= NewStringMessageHandler()
	h.print = false

	habitats:=make([]*Habitat,numOfHabitats)
	for i:=0;i<len(habitats);i++ {
		h,e:=NewHabitat(h)
		if e!=nil {
			t.Fail()

		}
		habitats[i]=h
		Info("Habitat HabitatID:"+habitats[i].ServiceID().String())
	}

	time.Sleep(time.Second*2)
	for i:=1;i<len(habitats)-1;i++ {
		go sendScale(h, habitats[i], habitats[i+1], numOfMessages)
	}

	time.Sleep(time.Second*2)

	if h.replyCount!=numOfMessages {
		t.Fail()
		Error("Expected "+strconv.Itoa(numOfMessages)+" and got "+strconv.Itoa(h.replyCount))
	} else {
		Info("Passed sending & receiving "+strconv.Itoa(h.replyCount)+ " messages")
	}

	for _,hb:=range habitats {
		hb.Shutdown()
	}
	time.Sleep(time.Second*2)
}

func TestHabitatAndMessageScale(t *testing.T) {
	MTU = 512
	ENCRYPTED=false
	numOfMessages:=10000
	numOfHabitats:=50

	h:= NewStringMessageHandler()
	h.print = false

	habitats:=make([]*Habitat,numOfHabitats)
	for i:=0;i<len(habitats);i++ {
		h,e:=NewHabitat(h)
		if e!=nil {
			t.Fail()

		}
		habitats[i]=h
		Info("Habitat HabitatID:"+habitats[i].ServiceID().String())
	}

	time.Sleep(time.Second*4)
	for i:=1;i<len(habitats)-1;i++ {
		go sendScale(h, habitats[i], habitats[i+1], numOfMessages)
	}

	time.Sleep(time.Second*10)

	if h.replyCount!=numOfMessages*(numOfHabitats-2) {
		t.Fail()
		Error("Expected "+strconv.Itoa(numOfMessages*(numOfHabitats-2))+" and got "+strconv.Itoa(h.replyCount))
	} else {
		Info("Passed sending & receiving "+strconv.Itoa(h.replyCount)+ " messages")
	}

	for _,hb:=range habitats {
		hb.Shutdown()
	}
	time.Sleep(time.Second*2)
}

func sendScale(h *StringMessageHandler, h1,h2 *Habitat, size int) {
	for i:=0;i<size;i++ {
		h.SendString("Hello World:"+strconv.Itoa(i),h1,h2.ServiceID())
	}
}

func TestPublish(t *testing.T) {
	MTU = 512
	ENCRYPTED=false
	numOfHabitats:=50

	h:= NewStringMessageHandler()

	habitats:=make([]*Habitat,numOfHabitats)
	for i:=0;i<len(habitats);i++ {
		h,e:=NewHabitat(h)
		if e!=nil {
			t.Fail()

		}
		habitats[i]=h
		Info("Habitat HabitatID:"+habitats[i].ServiceID().String())
	}

	time.Sleep(time.Second*2)

	publishID:=NewServiceID(PUBLISH_HID,0,"publish")

	h.SendString("Hello World Multicast",habitats[2],publishID)

	time.Sleep(time.Second*3)

	if h.replyCount!=len(habitats) {
		t.Fail()
		Error("Expected "+strconv.Itoa(len(habitats))+" and got "+strconv.Itoa(h.replyCount))
	} else {
		Info("Passed sending & receiving "+strconv.Itoa(h.replyCount)+ " messages")
	}

	for _,hb:=range habitats {
		hb.Shutdown()
	}
	time.Sleep(time.Second*2)
}

func TestHabitatEncrypted(t *testing.T) {
	MTU = 512
	ENCRYPTED=true
	h:= NewStringMessageHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	Info("Node1:",n1.ServiceID().String()," Node2:",n2.ServiceID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.ServiceID())
	h.SendString("Hello World",n2,n1.ServiceID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
	n1.Shutdown()
	n2.Shutdown()
	time.Sleep(time.Second*2)
}

func TestHabitatAndMessageScaleSecure(t *testing.T) {
	MTU = 512
	ENCRYPTED=true
	numOfMessages:=10000
	numOfHabitats:=50

	h:= NewStringMessageHandler()
	h.print = false

	habitats:=make([]*Habitat,numOfHabitats)
	for i:=0;i<len(habitats);i++ {
		h,e:=NewHabitat(h)
		if e!=nil {
			t.Fail()

		}
		habitats[i]=h
		Info("Habitat HabitatID:"+habitats[i].ServiceID().String())
	}

	time.Sleep(time.Second*6)

	for i:=1;i<len(habitats)-1;i++ {
		go sendScale(h, habitats[i], habitats[i+1], numOfMessages)
	}

	time.Sleep(time.Second*10)

	if h.replyCount!=numOfMessages*(numOfHabitats-2) {
		t.Fail()
		Error("Expected "+strconv.Itoa(numOfMessages*(numOfHabitats-2))+" and got "+strconv.Itoa(h.replyCount))
	} else {
		Info("Passed sending & receiving "+strconv.Itoa(h.replyCount)+ " messages")
	}

	for _,hb:=range habitats {
		hb.Shutdown()
	}
	time.Sleep(time.Second*2)
}

func TestDestinationUnreachable(t *testing.T) {
	MTU = 512
	ENCRYPTED=false
	h:= NewStringMessageHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	Info("Node1:",n1.ServiceID().String()," Node2:",n2.ServiceID().String())

	time.Sleep(time.Second*2)

	unreachInternal:=NewServiceID(NewHID(GetLocalIpAddress(),52005),0,"")

	h.SendString("Hello World",n2,unreachInternal)
	h.SendString("Hello World",n1,unreachInternal)

	time.Sleep(time.Second*2)

	if h.unreachCount!=2 {
		t.Fail()
		Error("Expected 2 and got "+strconv.Itoa(h.unreachCount))
	}
	n1.Shutdown()
	n2.Shutdown()
	time.Sleep(time.Second*2)
}

func TestMultiPartUnreachable(t *testing.T) {
	MTU = 4
	ENCRYPTED=false
	h:= NewStringMessageHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	Info("Node1:",n1.ServiceID().String()," Node2:",n2.ServiceID().String())

	time.Sleep(time.Second*2)
	unreachInternal:=NewServiceID(NewHID(GetLocalIpAddress(),52005),0,"")
	h.SendString("Hello World",n2,unreachInternal)

	time.Sleep(time.Second*2)

	if h.unreachCount!=1 {
		t.Fail()
		Error("Expected 1 and got "+strconv.Itoa(h.unreachCount))
	}
	n1.Shutdown()
	n2.Shutdown()
	time.Sleep(time.Second*2)
}

func TestShutdown(t *testing.T) {
	h:= NewStringMessageHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		Error(e)
		return
	}
	n2,e:=NewHabitat(h)

	time.Sleep(time.Second*3)

	n1.Shutdown()
	n2.Shutdown()

	time.Sleep(time.Second*5)
}