package main

import (
	"os"
	"fmt"
	"log"
	"flag"
	"bytes"
	"strings"
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

func (m *LoadMessage) Dump(w io.Writer) {
	fmt.Fprintln(w, "timestamp:", m.Timestamp)
	fmt.Fprintln(w, "interval:", m.Interval)
	fmt.Fprintf(w, "uptime: %.2f %.2f\n", m.Proc_load.Uptime_total, m.Proc_load.Uptime_idle)
	fmt.Fprintf(w, "loadavg: %.2f %.2f %.2f\n", m.Proc_load.Loadavg[0], m.Proc_load.Loadavg[1], m.Proc_load.Loadavg[2])
	fmt.Fprintf(w, "procs: all %d, running %d, iowait %d, zombie %d\n", m.Proc_load.Procs_all,
		m.Proc_load.Procs_running, m.Proc_load.Procs_iowait, m.Proc_load.Procs_zombie)

	for i := 0; i < len(m.Cpu_load.Items); i ++ {
		fmt.Fprintf(w, "cpu%d: user %d, sys %d, iowait %d, idle %d\n", i,
			m.Cpu_load.Items[i].Rate_user, m.Cpu_load.Items[i].Rate_sys,
			m.Cpu_load.Items[i].Rate_iowait, m.Cpu_load.Items[i].Rate_idle)
	}
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
		fmt.Println()
		lm.Dump(os.Stdout)
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
			fmt.Fprintln(peer.logfile)
			fmt.Fprintln(peer.logfile, "LoadMessage from", addr.IP)
			lm.Dump(peer.logfile)
			break
		}
	}
}

func main() {
	var err error
	var peers []LoadPeer

	f_server := flag.Bool("l", false, "server mode")
	f_port := flag.Int("p", 9999, "udp port to listen and (default) send")
	f_peers := flag.String("P", "", "peers, comma separated ipaddr[:port]")
	f_monitor := flag.Bool("m", true, "monitor local computer")
	flag.Parse()

	if len(*f_peers) > 0 {
		s := strings.Split(*f_peers, ",")
		peers = make([]LoadPeer, len(s))
		for i,ss := range s {
			if strings.Index(ss, ":") < 0 { ss = fmt.Sprintf("%s:%d", ss, *f_port) }
			peers[i].addr,err = net.ResolveUDPAddr("udp", ss)
			if err != nil { log.Fatal("invalid peer address:", ss) }
			peers[i].ipaddr = &peers[i].addr.IP
			peers[i].conn,err = net.DialUDP("udp", nil, peers[i].addr)
			if err != nil { log.Fatal("failed connect to udp:", peers[i].addr) }
			if *f_server {
				filename := fmt.Sprintf("%s.log", peers[i].addr.IP)
				peers[i].logfile,err = os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
				if err != nil { log.Fatal("failed open logfile:", filename) }
			}
		}
	}

	if *f_server {
		go Receiver(*f_port, peers)
	}

	if *f_monitor {
		Sender(1, nil, peers)
	} else {
		for { time.Sleep(1 * time.Second) }
	}
}
