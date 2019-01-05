package habitat

import (
	"bytes"
	"strconv"
	"sync"
)

type InterfaceStatistics struct {
	txMessages int64
	rxMessages int64
	txPackets  int64
	rxPackets  int64
	txBytes    int64
	rxBytes    int64
	txTime     int64
	rxTime     int64
	mtx  *sync.Mutex
}

func newInterfaceStatistics () *InterfaceStatistics {
	ist:=&InterfaceStatistics{}
	ist.mtx = &sync.Mutex{}
	return ist
}

func (ist *InterfaceStatistics) AddTxMessages(){
	ist.mtx.Lock()
	defer ist.mtx.Unlock()
	ist.txMessages++
}

func (ist *InterfaceStatistics) AddRxMessages(){
	ist.mtx.Lock()
	defer ist.mtx.Unlock()
	ist.rxMessages++
}

func (ist *InterfaceStatistics) AddTxPackets(data []byte){
	ist.txPackets++
	ist.txBytes+=int64(len(data))
}

func (ist *InterfaceStatistics) AddRxPackets(data []byte){
	ist.rxPackets++
	ist.rxBytes+=int64(len(data))
}

func (ist *InterfaceStatistics) AddTxTime(t int64) {
	ist.txTime+=t
}

func (ist *InterfaceStatistics) AddTxTimeSync(t int64) {
	ist.mtx.Lock()
	defer ist.mtx.Unlock()
	ist.txTime+=t
}

func (ist *InterfaceStatistics) String() string {
	buff:=&bytes.Buffer{}
	buff.WriteString("Rx Messages:"+strconv.Itoa(int(ist.rxMessages)))
	buff.WriteString(" Tx Messages:"+strconv.Itoa(int(ist.txMessages)))
	buff.WriteString(" Rx Packets:"+strconv.Itoa(int(ist.rxPackets)))
	buff.WriteString(" Tx Packets:"+strconv.Itoa(int(ist.txPackets)))
	buff.WriteString(" Rx Bytes:"+strconv.Itoa(int(ist.rxBytes)))
	buff.WriteString(" Tx Bytes:"+strconv.Itoa(int(ist.txBytes)))
	buff.WriteString(" Avg Tx Speed:"+ist.getTxSpeed())
	return buff.String()
}

func (ist *InterfaceStatistics) getTxSpeed() string {
	timeFloat:=float64(ist.txTime/1000000000)
	if int64(timeFloat)==0 {
		//not enought data
		return "N/A"
	}
	speed:=float64(ist.txBytes)/timeFloat

	if int64(speed)/1024==0 {
		return strconv.Itoa(int(speed))+" Bytes/Sec"
	}
	speed=speed/1024
	if int64(speed)/1024==0 {
		return strconv.Itoa(int(speed))+" Kilo Bytes/Sec"
	}
	speed=speed/1024
	s:=strconv.FormatFloat(speed, 'f', 2, 64)
	return s+" Mega Bytes/Sec"
}
