package tests

import (
	. "github.com/saichler/habitat"
	log "github.com/sirupsen/logrus"
	"testing"
	"time"
)

func TestNodeStart(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	h:=&StringFrameHandler{}
	n1,e:=NewHabitat(h)
	if e!=nil {
		log.Error(e)
		return
	}
	n1.Start()

	n2,e:=NewHabitat(h)
	n2.Start()

	log.Info("Node1:",n1.GetNID().FormattedString()," Node2:",n2.GetNID().FormattedString())

	time.Sleep(time.Second*10)

	h.SendString("Hello World",n1,n2.GetNID())
	h.SendString("Hello World",n2,n1.GetNID())

	time.Sleep(time.Second*10)

}
