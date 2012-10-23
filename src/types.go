package main

import (
	"io"
	"bytes"
)

const (
	MessageVersion = 1
	// subpacket codes
	SPC_ProcLoad = 10
	SPC_CPULoad = 11
	SPC_MemoryLoad = 12
	SPC_IOLoad = 13
	SPC_NetworkLoad = 14
)

type Subpacket interface {
	Encode() (spcode uint8, buf *bytes.Buffer)
	Decode(splen uint8, r io.Reader) error
}

type ProcLoad struct {
	Uptime_total, Uptime_idle float32
	Loadavg [3]float32
	Procs_all, Procs_running, Procs_iowait, Procs_zombie int32
}

type CPUItem struct {
	// normalize to 0-255, stands for 0-100%
	Rate_user, Rate_sys, Rate_iowait, Rate_idle uint8
}

type CPULoad struct {
	Interval float32
	Items []CPUItem
	Current [][4]int64
}

type MemoryLoad struct {
	free, buffers, cached, dirty, active, swapcached, swaptotal, swapfree uint32
}

type DiskItem struct {
	blk_read, blk_written, byte_read, byte_written uint32
}

type IOLoad struct {
	Items []DiskItem
	Current [][4]int64
}

type LoadMessage struct {
	Interval uint16
	Timestamp uint32 // timestamp: seconds from 2010-01-01 00:00:00

	Proc_load ProcLoad
	Cpu_load CPULoad
	Mem_load MemoryLoad
	Io_load IOLoad
	//Net_load NetworkLoad
}
