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
	avgSpeed   int64
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
	ist.mtx.Lock()
	defer ist.mtx.Unlock()
	ist.rxPackets++
	ist.rxBytes+=int64(len(data))
}

func (ist *InterfaceStatistics) SetSpeed(speed int64) {
	ist.mtx.Lock()
	defer ist.mtx.Unlock()
	if ist.avgSpeed==0 || speed<ist.avgSpeed{
		ist.avgSpeed = speed
	}
}

func (ist *InterfaceStatistics) String() string {
	buff:=&bytes.Buffer{}
	buff.WriteString("Rx Messages:"+strconv.Itoa(int(ist.rxMessages)))
	buff.WriteString(" Tx Messages:"+strconv.Itoa(int(ist.txMessages)))
	buff.WriteString(" Rx Packets:"+strconv.Itoa(int(ist.rxPackets)))
	buff.WriteString(" Tx Packets:"+strconv.Itoa(int(ist.txPackets)))
	buff.WriteString(" Rx Bytes:"+strconv.Itoa(int(ist.rxBytes)))
	buff.WriteString(" Tx Bytes:"+strconv.Itoa(int(ist.txBytes)))
	buff.WriteString(" Avg Speed:"+ist.getSpeed())
	return buff.String()
}

func (ist *InterfaceStatistics) getSpeed() string {
	speed:=float64(ist.avgSpeed)
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
