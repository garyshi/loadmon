package main

import (
	"io"
	"os"
	"fmt"
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

func cpuload_convtoint(fields []string) [4]int64 {
	var i1, i2, i3, i4 int64
	i1, _ = strconv.ParseInt(fields[0], 0, 64)
	i2, _ = strconv.ParseInt(fields[2], 0, 64)
	i3, _ = strconv.ParseInt(fields[4], 0, 64)
	i4, _ = strconv.ParseInt(fields[3], 0, 64)
	return [4]int64{i1, i2, i3, i4}
}

func cpuload_getstat() (rslt [][4]int64, err error) {
	var line string
	var fields []string

	file, err := os.Open("/proc/stat")
	if err != nil {
		fmt.Println(fmt.Errorf("CPULoad.Probe: failed optn /proc/stat"))
		return
	}
	defer file.Close()

	rslt = make([][4]int64, 0)
	reader := bufio.NewReader(file)

	line, err = reader.ReadString('\n')
	for err == nil {
		fields = strings.Fields(line)
		if strings.HasPrefix(fields[0], "cpu") {
			rslt = append(rslt, cpuload_convtoint(fields[1:]))
		}
		line, err = reader.ReadString('\n') 
	}

	if err == io.EOF { err = nil }
	return
}

func (load *CPULoad) ProbeInit() (err error) {
	rslt, err := cpuload_getstat()
	load.Current = rslt
	load.Items = make([]CPUItem, len(rslt))
	return err
}

func (load *CPULoad) Probe() (err error) {
	var all float32
	var diff [4]float32

	rslt, err := cpuload_getstat()
	if len(load.Current) != len(rslt) {
		return errors.New("different CPU numbers")
	}

	for i := 0; i < len(load.Current); i++ {
		all = 0
		for j := 0; j < 4; j++ {
			diff[j] = float32(rslt[i][j] - load.Current[i][j])
			all += diff[j]
		}
		load.Items[i].Rate_user = uint8(diff[0] / all * 255)
		load.Items[i].Rate_sys = uint8(diff[1] / all * 255)
		load.Items[i].Rate_iowait = uint8(diff[2] / all * 255)
		load.Items[i].Rate_idle = uint8(diff[3] / all * 255)
	}

	load.Current = rslt
	return nil
}

func (load *MemoryLoad) Probe() (err error) {
	var value int64
	var line string
	var fields []string

	file, err := os.Open("/proc/meminfo")
	if err != nil {
		fmt.Println(fmt.Errorf("MemoryLoad.Probe: failed optn /proc/meminfo"))
		return
	}
	defer file.Close()
	reader := bufio.NewReader(file)

	line, err = reader.ReadString('\n')
	for err == nil {
		fields = strings.Split(line, ":")
		fields[1] = strings.Trim(fields[1], " kbB\r\n")
		value, err = strconv.ParseInt(fields[1], 0, 32)
		if err != nil { return }
		switch{
		case fields[0] == "MemFree": load.free = uint32(value)
		case fields[0] == "Buffers": load.buffers = uint32(value)
		case fields[0] == "Cached": load.cached = uint32(value)
		case fields[0] == "Dirty": load.dirty = uint32(value)
		case fields[0] == "Active": load.active = uint32(value)
		case fields[0] == "SwapCached": load.swapcached = uint32(value)
		case fields[0] == "SwapTotal": load.swaptotal = uint32(value)
		case fields[0] == "SwapFree": load.swapfree = uint32(value)
		}
		line, err = reader.ReadString('\n') 
	}

	if err == io.EOF { err = nil }
	return
}

func ioload_convtoint(fields []string) [4]int64 {
	var i1, i2, i3, i4 int64

	i1, _ = strconv.ParseInt(fields[3], 0, 64)
	i2, _ = strconv.ParseInt(fields[5], 0, 64)
	i3, _ = strconv.ParseInt(fields[7], 0, 64)
	i4, _ = strconv.ParseInt(fields[9], 0, 64)
	return [4]int64{i1, i2 * 512, i3, i4 * 512}
}

func ioload_getstat() (rslt [][4]int64, names []string, err error) {
	var line string
	var fields []string

	file, err := os.Open("/proc/diskstats")
	if err != nil {
		fmt.Println(fmt.Errorf("IOLoad.Probe: failed optn /proc/diskstats"))
		return
	}
	defer file.Close()

	rslt = make([][4]int64, 0)
	reader := bufio.NewReader(file)

	line, err = reader.ReadString('\n')
	for err == nil {
		fields = strings.Fields(line)
		if strings.HasPrefix(fields[2], "sd") && len(fields[2]) == 3 {
			rslt = append(rslt, ioload_convtoint(fields))
			names = append(names, fields[2])
		}
		line, err = reader.ReadString('\n') 
	}

	if err == io.EOF { err = nil }
	return
}

func (load *IOLoad) ProbeInit() (err error) {
	rslt, names, err := ioload_getstat()
	load.Current = rslt;
	load.names = names
	load.Items = make([]DiskItem, len(rslt))
	return
}

func (load *IOLoad) Probe() (err error) {
	rslt, names, err := ioload_getstat()
	if len(load.Current) != len(rslt) {
		return errors.New("different sdX numbers")
	}

	for i := 0; i < len(load.Current); i++ {
		load.Items[i].blk_read = uint32(rslt[i][0] - load.Current[i][0])
		load.Items[i].byte_read = uint32(rslt[i][1] - load.Current[i][1])
		load.Items[i].blk_written = uint32(rslt[i][2] - load.Current[i][2])
		load.Items[i].byte_written = uint32(rslt[i][3] - load.Current[i][3])
		load.Items[i].name = names[i]
	}

	load.Current = rslt
	return nil
}

func (m *LoadMessage) ProbeInit() error {
	m.Cpu_load.ProbeInit()
	m.Io_load.ProbeInit()
	return nil
}

func (m *LoadMessage) ProbeRotate() error {
	return nil
}

func (m *LoadMessage) Probe() error {
	m.Timestamp = GetTimestamp()

	if err := m.Proc_load.Probe(); err != nil { return err }
	if err := m.Cpu_load.Probe(); err != nil { return err }
	if err := m.Mem_load.Probe(); err != nil { return err }
	if err := m.Io_load.Probe(); err != nil { return err }

	return nil
}
