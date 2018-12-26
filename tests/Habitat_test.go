package tests

import (
	. "github.com/saichler/habitat"
	log "github.com/sirupsen/logrus"
	"os"
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

func TestNode(t *testing.T) {
	MTU = 512
	h:=&StringFrameHandler{}

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
}

func TestNodeMultiPart(t *testing.T) {
	MTU = 4
	h:=&StringFrameHandler{}

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
}