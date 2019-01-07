package tests

import (
	. "github.com/saichler/habitat/service"
	. "github.com/saichler/utils/golang"
	"os"
	"testing"
	"time"
)


func setup(){}
func tearDown(){}

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	tearDown()
	os.Exit(code)
}


func TestShutdown(t *testing.T){
	svr,err:=NewServiceManager()
	if err!=nil {
		t.Fatal()
		Error(err)
		return
	}

	time.Sleep(time.Second)

	svr.Shutdown()

	time.Sleep(time.Second*10)
}