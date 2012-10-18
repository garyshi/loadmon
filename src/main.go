package main

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"time"
	"io"
	"net"
	"encoding/hex"
)

type LoadPeer struct {
	ipaddr *net.IP
	addr *net.UDPAddr
	conn *net.UDPConn
	logfile *os.File
}

func (m *LoadMessage) Dump() {
	fmt.Println("timestamp:", m.Timestamp)
	fmt.Println("interval:", m.Interval)
	fmt.Printf("uptime: %.2f %.2f\n", m.Proc_load.Uptime_total, m.Proc_load.Uptime_idle)
	fmt.Printf("loadavg: %.2f %.2f %.2f\n", m.Proc_load.Loadavg[0], m.Proc_load.Loadavg[1], m.Proc_load.Loadavg[2])
	fmt.Printf("procs: all %d, running %d, iowait %d, zombie %d\n", m.Proc_load.Procs_all,
		m.Proc_load.Procs_running, m.Proc_load.Procs_iowait, m.Proc_load.Procs_zombie)

/*
	for i := 0; i < len(m.Cpu_load.Items); i ++ {
		fmt.Printf("cpu%d: user %d, sys %d, iowait %d, idle %d\n", i,
			m.Cpu_load.Items[i].Rate_user, m.Cpu_load.Items[i].Rate_sys,
			m.Cpu_load.Items[i].Rate_iowait, m.Cpu_load.Items[i].Rate_idle)
	}
*/
}

func Sender(interval int, logfile io.Writer, peers []LoadPeer) {
	lm := LoadMessage{Interval:uint16(interval)}
	lm.ProbeInit()

	for {
		var buffer bytes.Buffer

		time.Sleep(time.Duration(lm.Interval) * time.Second)

		lm.Probe()
		if err := lm.Encode(&buffer); err != nil {
			log.Fatal("encode error:", err)
		}

		fmt.Println()
		fmt.Println("Local LoadMessage")
		dumper := hex.Dumper(os.Stdout)
		dumper.Write(buffer.Bytes())
		lm.Dump()
		//lm.WriteLog(logfile)

		for _,peer := range peers {
			peer.conn.Write(buffer.Bytes())
		}
		lm.ProbeRotate()
	}
}

func Receiver(port int, peers []LoadPeer) {
	var lm LoadMessage

	buf := make([]byte, 2000) // max should be 1500
	conn,err := net.ListenUDP("udp", &net.UDPAddr{Port:port})
	if err != nil {
		fmt.Println("Failed listen UDP")
		return
	}

	for {
		n,addr,err := conn.ReadFromUDP(buf)
		if err != nil { continue }
		if n == len(buf) { fmt.Println("Warning: received very long packet") }

		for _,peer := range peers {
			if !addr.IP.Equal(*peer.ipaddr) { continue }
			err = lm.Decode(bytes.NewReader(buf[:n]))
			if err != nil { fmt.Println("Error decode packet:", err) }
			//lm.WriteLog(peer.logfile)
			fmt.Println("LoadMessage from", addr.IP)
			lm.Dump()
			fmt.Println()
			break
		}
	}
}

func main() {
	var err error
	var buffer bytes.Buffer

	lm1 := &LoadMessage{Timestamp:100, Interval:10}
	lm1.Proc_load.Probe()

	lm1.Cpu_load.Items = make([]CPUItem, 3)
	lm1.Cpu_load.Items[0] = CPUItem{255*3/10, 255*2/10, 255*2/10, 255*3/10}
	lm1.Cpu_load.Items[1] = CPUItem{255*4/10, 255*1/10, 255*1/10, 255*4/10}
	lm1.Cpu_load.Items[2] = CPUItem{255*2/10, 255*2/10, 255*3/10, 255*3/10}

	if err = lm1.Encode(&buffer); err != nil {
		log.Fatal("encode error:", err)
	}

	dumper := hex.Dumper(os.Stdout)
	fmt.Println("message size:", buffer.Len())
	dumper.Write(buffer.Bytes())
	fmt.Println()
	fmt.Println()

	var lm2 LoadMessage
	if err = lm2.Decode(&buffer); err != nil {
		log.Fatal("decode error:", err)
	}

	lm2.Dump()

	Sender(1, nil, nil)
}
