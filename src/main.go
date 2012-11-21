package main

import (
	"io"
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
		if *f_verbose > 1 {
			fmt.Printf("Local LoadMessage, size=%d\n", buffer.Len())
			hex.Dumper(os.Stdout).Write(buffer.Bytes())
			fmt.Println()
		}
		lm.Dump(os.Stdout)
		if logfile != nil { logfile.WriteMessage(buffer.Bytes()) }

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
			if peer.logfile != nil { peer.logfile.WriteMessage(buf[:n]) }
			break
		}
	}
}

func DumpLogFile(filename string) {
	var lm LoadMessage

	logfile,err := OpenLogFile(filename, MODE_READ)
	if err != nil {
		fmt.Println("Error open log file:", err)
		return
	}

	for {
		ts,buffer,err := logfile.ReadMessage()
		if err != nil { break }
		t := FromTimestamp(ts).Format("20060102-150405")
		err = lm.Decode(bytes.NewReader(buffer))
		if err != nil {
			fmt.Println("Error decode packet:", err)
			break
		}

		fmt.Println()
		if *f_verbose > 1 {
			fmt.Printf("Read LoadMessage, local time %s, size=%d\n", t, len(buffer))
			hex.Dumper(os.Stdout).Write(buffer)
			fmt.Println()
		}
		lm.Dump(os.Stdout)
	}

	if err != nil && err != io.EOF { fmt.Println(err) }
}

var f_readfile = flag.String("r", "", "decode log file")
var f_server = flag.Bool("l", false, "server mode")
var f_port = flag.Int("p", 9999, "udp port to listen and (default) send")
var f_peers = flag.String("P", "", "peers, comma separated ipaddr[:port]")
var f_monitor = flag.Bool("m", true, "monitor local computer")
var f_nolog = flag.Bool("n", true, "don't write log files on this node")
var f_interval = flag.Int("i", 10, "local monitor interval")
var f_verbose = flag.Int("v", 1, "verbose level")

func main() {
	var err error
	var peers []LoadPeer

	now := time.Now()
	flag.Parse()

	if *f_readfile != "" {
		DumpLogFile(*f_readfile)
		return
	}

	if len(*f_peers) > 0 {
		s := strings.Split(*f_peers, ",")
		peers = make([]LoadPeer, len(s))
		for i,ss := range s {
			if strings.Index(ss, ":") < 0 { ss = fmt.Sprintf("%s:%d", ss, *f_port) }
			peers[i].addr,err = net.ResolveUDPAddr("udp", ss)
			if err != nil { log.Fatal("invalid peer address:", ss) }
			peers[i].conn,err = net.DialUDP("udp", nil, peers[i].addr)
			if err != nil { log.Fatal("failed connect to udp:", peers[i].addr) }
			if *f_server && !*f_nolog {
				peers[i].logfile,err = OpenRotateLogFile(peers[i].addr.IP.String(), &now, MODE_APPEND)
				if err != nil { log.Fatal("failed open logfile:", err) }
			}
		}
	}

	if *f_server {
		go Receiver(*f_port, peers)
	}

	if *f_monitor {
		var logfile *LogFile
		hostname,err := os.Hostname()
		if err != nil { hostname = "localhost" }
		if !*f_nolog { logfile,err = OpenRotateLogFile(hostname, &now, MODE_APPEND) }
		Sender(*f_interval, logfile, peers)
	} else {
		for { time.Sleep(1 * time.Second) }
	}
}
