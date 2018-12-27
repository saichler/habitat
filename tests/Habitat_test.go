package tests

import (
	. "github.com/saichler/habitat"
	log "github.com/sirupsen/logrus"
	"os"
	"strconv"
	"testing"
	"time"
)

func setup(){
	log.SetLevel(log.DebugLevel)
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
	h:=NewStringFrameHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		log.Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	log.Info("Node1:",n1.GetNID().String()," Node2:",n2.GetNID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.GetNID())
	h.SendString("Hello World",n2,n1.GetNID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		log.Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
}

func TestSwitch(t *testing.T) {
	MTU = 512
	h:=NewStringFrameHandler()

	_,e:=NewHabitat(h)
	if e!=nil {
		log.Error(e)
		return
	}

	n1,e:=NewHabitat(h)
	n2,e:=NewHabitat(h)

	log.Info("Node1:",n1.GetNID().String()," Node2:",n2.GetNID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.GetNID())
	h.SendString("Hello World",n2,n1.GetNID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		log.Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
}

func TestMultiPart(t *testing.T) {
	MTU = 4
	h:=NewStringFrameHandler()

	n1,e:=NewHabitat(h)
	if e!=nil {
		log.Error(e)
		return
	}

	n2,e:=NewHabitat(h)

	log.Info("Node1:",n1.GetNID().String()," Node2:",n2.GetNID().String())

	time.Sleep(time.Second*2)

	h.SendString("Hello World",n1,n2.GetNID())
	h.SendString("Hello World",n2,n1.GetNID())

	time.Sleep(time.Second*2)

	if h.replyCount!=2 {
		t.Fail()
		log.Error("Expected 2 and got "+strconv.Itoa(h.replyCount))
	}
}

func TestMessageScale(t *testing.T) {
	MTU = 512
	numOfMessages:=10000
	numOfHabitats:=3

	h:=NewStringFrameHandler()
	h.print = false

	habitats:=make([]*Habitat,numOfHabitats)
	for i:=0;i<len(habitats);i++ {
		h,e:=NewHabitat(h)
		if e!=nil {
			t.Fail()

		}
		habitats[i]=h
		log.Info("Habitat HID:"+habitats[i].GetNID().String())
	}

	time.Sleep(time.Second*2)
	for i:=1;i<len(habitats)-1;i++ {
		go sendScale(h, habitats[i], habitats[i+1], numOfMessages)
	}

	time.Sleep(time.Second*2)

	if h.replyCount!=numOfMessages {
		t.Fail()
		log.Error("Expected "+strconv.Itoa(numOfMessages)+" and got "+strconv.Itoa(h.replyCount))
	} else {
		log.Info("Passed sending & receiving "+strconv.Itoa(h.replyCount)+ " messages")
	}
}

func TestHabitatAndMessageScale(t *testing.T) {
	MTU = 512
	numOfMessages:=10000
	numOfHabitats:=50

	h:=NewStringFrameHandler()
	h.print = false

	habitats:=make([]*Habitat,numOfHabitats)
	for i:=0;i<len(habitats);i++ {
		h,e:=NewHabitat(h)
		if e!=nil {
			t.Fail()

		}
		habitats[i]=h
		log.Info("Habitat HID:"+habitats[i].GetNID().String())
	}

	time.Sleep(time.Second*2)
	for i:=1;i<len(habitats)-1;i++ {
		go sendScale(h, habitats[i], habitats[i+1], numOfMessages)
	}

	time.Sleep(time.Second*5)

	if h.replyCount!=numOfMessages*(numOfHabitats-2) {
		t.Fail()
		log.Error("Expected "+strconv.Itoa(numOfMessages*(numOfHabitats-2))+" and got "+strconv.Itoa(h.replyCount))
	} else {
		log.Info("Passed sending & receiving "+strconv.Itoa(h.replyCount)+ " messages")
	}
}

func sendScale(h *StringFrameHandler, h1,h2 *Habitat, size int) {
	for i:=0;i<size;i++ {
		h.SendString("Hello World:"+strconv.Itoa(i),h1,h2.GetNID())
	}
}