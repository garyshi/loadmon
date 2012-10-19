package main

import (
	"os"
	"fmt"
	"time"
	"bytes"
	"bufio"
	"errors"
	"strconv"
	"strings"
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

func convtoint(fields []string) [4]int64 {
	var rslt []int64
	for _, field := range fields {
		i, _ := strconv.ParseInt(field, 0, 64)
		rslt = append(rslt, i)
	}
	return [4]int64{rslt[0], rslt[2], rslt[4], rslt[3]}
}

func (load *CPULoad) getstat() error {
	var err error
	var line string
	var fields []string

	file, err := os.Open("/proc/stat")
	if err != nil {
		fmt.Println(fmt.Errorf("CPULoad.Probe: failed optn /proc/stat"))
		return err
	}
	reader := bufio.NewReader(file)

	load.Current = [][4]int64{}
	line, err = reader.ReadString('\n')
	for err == nil {
		fields = strings.Fields(line)
		if strings.HasPrefix(fields[0], "cpu") && fields[0] != "cpu" {
			load.Current = append(load.Current, convtoint(fields[1:]))
		}
		line, err = reader.ReadString('\n') 
	}
	file.Close()

	return nil
}

func (load *CPULoad) renew(load_next *CPULoad) error {
	var all float32
	var diff [4]float32

	if len(load.Current) != len(load_next.Current) {
		return errors.New("different CPU numbers")
	}
	load.Items = make([]CPUItem, len(load.Current))

	for i := 0; i < len(load.Current); i++ {
		all = 0
		for j := 0; j < 4; j++ {
			diff[j] = float32(load_next.Current[i][j] - load.Current[i][j])
			all += diff[j]
		}
		load.Items[i].Rate_user = uint8(diff[0] / all * 255)
		load.Items[i].Rate_sys = uint8(diff[1] / all * 255)
		load.Items[i].Rate_iowait = uint8(diff[2] / all * 255)
		load.Items[i].Rate_idle = uint8(diff[3] / all * 255)
	}

	load.Current = load_next.Current
	return nil
}

func (load *CPULoad) ProbeInit() error {
	err := load.getstat()
	if err != nil {
		return err
	}
	return nil
}

func (load *CPULoad) Probe() error {
	var err error
	var load_cur CPULoad
	err = load_cur.getstat()
	if err != nil {
		return err
	}
	err = load.renew(&load_cur)
	if err != nil {
		return err
	}
	return nil
}

func (m *LoadMessage) ProbeInit() error {
	m.Cpu_load.ProbeInit()
	return nil
}

func (m *LoadMessage) ProbeRotate() error {
	return nil
}

func (m *LoadMessage) Probe() error {
	// seconds from 2000-01-01 00:00:00 UTC, can work till year 2136
	m.Timestamp = uint32(time.Now().Unix() - 946684800)

	if err := m.Proc_load.Probe(); err != nil { return err }
	if err := m.Cpu_load.Probe(); err != nil { return err }

	return nil
}
