package main

import (
	"fmt"
	"bytes"
	"strconv"
	"io/ioutil"
)

func (load *ProcLoad) Probe() error {
	var err error
	var buf []byte

	buf,err = ioutil.ReadFile("/proc/uptime")
	if err != nil {
		return fmt.Errorf("ProcLoad.Probe: failed open /proc/uptime");
	} else {
		fmt.Sscanf(string(buf), "%f %f", &load.Uptime_idle, &load.Uptime_total);
	}

	buf,err = ioutil.ReadFile("/proc/loadavg")
	if err != nil {
		return fmt.Errorf("ProcLoad.Probe: failed open /proc/loadavg");
	} else {
		fmt.Sscanf(string(buf), "%f %f %f", &load.Loadavg[0], &load.Loadavg[1], &load.Loadavg[2])
	}

	load.Procs_all = 0
	load.Procs_running = 0
	load.Procs_iowait = 0
	load.Procs_zombie = 0
	filelist,err := ioutil.ReadDir("/proc")
	for _,fileinfo := range filelist {
		pid,err := strconv.Atoi(fileinfo.Name())
		if err != nil { continue }
		buf,err = ioutil.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
		if err != nil { continue } // process may exit before we read the stat file
		i := bytes.LastIndex(buf, []byte(")"))
		if i < 0 || len(buf) < i+3 { continue }
		load.Procs_all ++
		switch buf[i+2] {
		case 'R': load.Procs_running ++
		case 'D': load.Procs_iowait ++
		case 'Z': load.Procs_zombie ++
		}
	}

	return nil
}
