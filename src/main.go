package main

import (
	"os"
	"net"
	"fmt"
	"log"
	"flag"
	"time"
	"bytes"
	"strings"
	"encoding/hex"
)

type LoadPeer struct {
	addr *net.UDPAddr
	conn *net.UDPConn
	logfile *LogFile
}

func Sender(interval int, logfile *LogFile, peers []LoadPeer) {
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
		fmt.Printf("Local LoadMessage, size=%d\n", buffer.Len())
		hex.Dumper(os.Stdout).Write(buffer.Bytes())
		fmt.Println()
		fmt.Println()
		lm.Dump(os.Stdout)
		logfile.WriteMessage(buffer.Bytes())

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
			if !addr.IP.Equal(peer.addr.IP) { continue }
			err = lm.Decode(bytes.NewReader(buf[:n]))
			if err != nil { fmt.Println("Error decode packet:", err) }
			peer.logfile.WriteMessage(buf[:n])
			break
		}
	}
}

func main() {
	var err error
	var peers []LoadPeer

	now := time.Now()
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
			peers[i].conn,err = net.DialUDP("udp", nil, peers[i].addr)
			if err != nil { log.Fatal("failed connect to udp:", peers[i].addr) }
			if *f_server {
				filename := LogFileName("", peers[i].addr.IP.String(), &now)
				peers[i].logfile,err = OpenLogFile(filename, MODE_APPEND)
				if err != nil { log.Fatal("failed open logfile:", filename) }
			}
		}
	}

	if *f_server {
		go Receiver(*f_port, peers)
	}

	if *f_monitor {
		hostname,err := os.Hostname()
		if err != nil { hostname = "localhost" }
		filename := LogFileName("", hostname, &now)
		logfile,err := OpenLogFile(filename, MODE_APPEND)
		Sender(1, logfile, peers)
	} else {
		for { time.Sleep(1 * time.Second) }
	}
}
