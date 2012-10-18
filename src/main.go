package main

import (
	"os"
	"fmt"
	"log"
	"bytes"
	"encoding/hex"
)

func main() {
	var err error
	var buffer bytes.Buffer

	lm1 := &LoadMessage{Timestamp:100, Interval:10}

	lm1.Proc_load.Uptime_total = 10000
	lm1.Proc_load.Uptime_idle = 8000
	lm1.Proc_load.Loadavg[0] = 1.0
	lm1.Proc_load.Loadavg[1] = 0.8
	lm1.Proc_load.Loadavg[2] = 0.5
	lm1.Proc_load.Procs_all = 100
	lm1.Proc_load.Procs_running = 5
	lm1.Proc_load.Procs_iowait = 2
	lm1.Proc_load.Procs_zombie = 0
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

	fmt.Println("timestamp:", lm2.Timestamp)
	fmt.Println("interval:", lm2.Interval)
	fmt.Printf("uptime: %.2f %.2f\n", lm2.Proc_load.Uptime_total, lm2.Proc_load.Uptime_idle)
	fmt.Printf("loadavg: %.2f %.2f %.2f\n", lm2.Proc_load.Loadavg[0], lm2.Proc_load.Loadavg[1], lm2.Proc_load.Loadavg[2])
	fmt.Printf("procs: all %d, running %d, iowait %d, zombie %d\n", lm2.Proc_load.Procs_all,
		lm2.Proc_load.Procs_running, lm2.Proc_load.Procs_iowait, lm2.Proc_load.Procs_zombie)
	for i := 0; i < len(lm2.Cpu_load.Items); i ++ {
		fmt.Printf("cpu%d: user %d, sys %d, iowait %d, idle %d\n", i,
			lm2.Cpu_load.Items[i].Rate_user, lm2.Cpu_load.Items[i].Rate_sys,
			lm2.Cpu_load.Items[i].Rate_iowait, lm2.Cpu_load.Items[i].Rate_idle)
	}
}
