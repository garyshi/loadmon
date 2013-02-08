package main

import (
	"os"
	"fmt"
	"bytes"
	"errors"
	"strconv"
	"strings"
	"io/ioutil"
	"./sutils"
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

func Fields2Int(fields []string, cols []int) (rslt []int64, err error) {
	var r int64
	for _, v := range cols {
		r, err = strconv.ParseInt(fields[v], 0, 64)
		if err != nil { return }
		rslt = append(rslt, r)
	}
	return
}

func cpuload_getstat() (rslt [][4]int64, err error) {
	file, err := os.Open("/proc/stat")
	if err != nil {
		fmt.Println(fmt.Errorf("CPULoad.Probe: failed open /proc/stat"))
		return
	}
	defer file.Close()

	sutils.ReadLines(file, func (line string) (err error) {
		fields := strings.Fields(line)
		if !strings.HasPrefix(fields[0], "cpu") { return }

		i, err := Fields2Int(fields, []int{1,3,5,4})
		if err != nil { return }
		rslt = append(rslt, [4]int64{i[0], i[1], i[2], i[3]})
		return
	})
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
		load.Items[i].Rate_user = uint8(diff[0] / all * 255 + 0.5)
		load.Items[i].Rate_sys = uint8(diff[1] / all * 255 + 0.5)
		load.Items[i].Rate_iowait = uint8(diff[2] / all * 255 + 0.5)
		load.Items[i].Rate_idle = uint8(diff[3] / all * 255 + 0.5)
	}

	load.Current = rslt
	return nil
}

func (load *MemoryLoad) Probe() (err error) {
	file, err := os.Open("/proc/meminfo")
	if err != nil {
		fmt.Println(fmt.Errorf("MemoryLoad.Probe: failed open /proc/meminfo"))
		return
	}
	defer file.Close()

	sutils.ReadLines(file, func (line string) (err error) {
		fields := strings.Split(line, ":")
		fields[1] = strings.Trim(fields[1], " kbB\r\n")
		value, err := strconv.ParseInt(fields[1], 0, 32)
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
		return
	})
	return
}

func ioload_getstat() (rslt [][4]int64, names []string, err error) {
	file, err := os.Open("/proc/diskstats")
	if err != nil {
		fmt.Println(fmt.Errorf("IOLoad.Probe: failed open /proc/diskstats"))
		return
	}
	defer file.Close()

	sutils.ReadLines(file, func (line string) (err error) {
		fields := strings.Fields(line)
		if !strings.HasPrefix(fields[2], "sd") || len(fields[2]) != 3 { return }

		i, err := Fields2Int(fields, []int{3,5,7,9})
		if err != nil { return }
		rslt = append(rslt, [4]int64{i[0], (i[1]+1)/2, i[2], (i[3]+1)/2})
		names = append(names, fields[2])
		return
	})
	return
}

func (load *IOLoad) ProbeInit() (err error) {
	rslt, _, err := ioload_getstat()
	load.Current = rslt
	load.Items = make([]DiskItem, len(rslt))
	return
}

func (load *IOLoad) Probe() (err error) {
	rslt, names, err := ioload_getstat()
	if len(load.Current) != len(rslt) {
		return errors.New("different sdX numbers")
	}

	for i := 0; i < len(load.Current); i++ {
		load.Items[i].tps_read = uint32(rslt[i][0] - load.Current[i][0])
		load.Items[i].kbytes_read = uint32(rslt[i][1] - load.Current[i][1])
		load.Items[i].tps_written = uint32(rslt[i][2] - load.Current[i][2])
		load.Items[i].kbytes_written = uint32(rslt[i][3] - load.Current[i][3])
		load.Items[i].name = names[i]
	}

	load.Current = rslt
	return nil
}

func network_getstat() (rslt [][4]int64, names []string, err error) {
	file, err := os.Open("/proc/net/dev")
	if err != nil {
		fmt.Println(fmt.Errorf("Network.Probe: failed open /proc/net/dev"))
		return
	}
	defer file.Close()

	sutils.ReadLines(file, func (line string) (err error) {
		if !strings.Contains(line, ":") { return }
		fields := strings.Fields(line)
		names = append(names, strings.Trim(fields[0], ":"))

		i, err := Fields2Int(fields, []int{1,2,9,10})
		if err != nil { return }
		rslt = append(rslt, [4]int64{(i[0]+1023)/1024, i[1], (i[2]+1023)/1024, i[3]})
		return
	})
	return
}

func (load *NetworkLoad) ProbeInit() (err error) {
	rslt, _, err := network_getstat()
	load.Current = rslt
	load.Items = make([]InterfaceItem, len(rslt))
	return
}

func (load *NetworkLoad) Probe() (err error) {
	rslt, names, err := network_getstat()
	if len(load.Current) != len(rslt) {
		return errors.New("different network numbers")
	}

	for i := 0; i < len(load.Current); i++ {
		load.Items[i].kbytes_read = uint32(rslt[i][0] - load.Current[i][0])
		load.Items[i].pkts_read = uint32(rslt[i][1] - load.Current[i][1])
		load.Items[i].kbytes_written = uint32(rslt[i][2] - load.Current[i][2])
		load.Items[i].pkts_written = uint32(rslt[i][3] - load.Current[i][3])
		load.Items[i].name = names[i]
	}

	load.Current = rslt
	return nil
}

func (m *LoadMessage) ProbeInit() error {
	m.Cpu_load.ProbeInit()
	m.Io_load.ProbeInit()
	m.Net_load.ProbeInit()
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
	if err := m.Net_load.Probe(); err != nil { return err }

	return nil
}
