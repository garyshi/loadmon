package main

import (
	"io"
	"fmt"
	"time"
)

func ToTimestamp(t time.Time) uint32 {
	// seconds from 2000-01-01 00:00:00 UTC, can work till year 2136
	ts := t.Unix() - 946684800
	return uint32(ts)
}

func FromTimestamp(ts uint32) time.Time {
	t := int64(ts) + 946684800
	return time.Unix(t, 0)
}

func GetTimestamp() uint32 {
	t := time.Now()
	return ToTimestamp(t)
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

	fmt.Fprintf(w, "mem: free %d, buffers %d, cached %d, dirty %d, active %d\n",
		m.Mem_load.free, m.Mem_load.buffers, m.Mem_load.cached,
		m.Mem_load.dirty, m.Mem_load.active)
	fmt.Fprintf(w, "mem swap: cached %d, total %d, free %d\n",
		m.Mem_load.swapcached, m.Mem_load.swaptotal, m.Mem_load.swapfree)

	for i := 0; i < len(m.Io_load.Items); i++ {
		fmt.Fprintf(w, "drv %s: blk_read %d, byte_read %d, blk_written %d, byte_written %d\n",
			m.Io_load.Items[i].name,
			m.Io_load.Items[i].blk_read, m.Io_load.Items[i].byte_read,
			m.Io_load.Items[i].blk_written, m.Io_load.Items[i].byte_written,
			)
	}
}
