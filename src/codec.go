package main

import (
	"io"
	"fmt"
	"bytes"
	"encoding/binary"
)

func (load *ProcLoad) Encode() (uint8, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, load.Uptime_total)
	binary.Write(buf, binary.BigEndian, load.Uptime_idle)
	binary.Write(buf, binary.BigEndian, load.Loadavg[0])
	binary.Write(buf, binary.BigEndian, load.Loadavg[1])
	binary.Write(buf, binary.BigEndian, load.Loadavg[2])
	binary.Write(buf, binary.BigEndian, load.Procs_all)
	binary.Write(buf, binary.BigEndian, load.Procs_running)
	binary.Write(buf, binary.BigEndian, load.Procs_iowait)
	binary.Write(buf, binary.BigEndian, load.Procs_zombie)
	return SPC_ProcLoad, buf
}

func (load *ProcLoad) Decode(splen uint8, r io.Reader) error {
	if splen != 36 { return fmt.Errorf("ProcLoad.Decode: invalid subpacket size (%d)", splen) }

	binary.Read(r, binary.BigEndian, &load.Uptime_total)
	binary.Read(r, binary.BigEndian, &load.Uptime_idle)
	binary.Read(r, binary.BigEndian, &load.Loadavg[0])
	binary.Read(r, binary.BigEndian, &load.Loadavg[1])
	binary.Read(r, binary.BigEndian, &load.Loadavg[2])
	binary.Read(r, binary.BigEndian, &load.Procs_all)
	binary.Read(r, binary.BigEndian, &load.Procs_running)
	binary.Read(r, binary.BigEndian, &load.Procs_iowait)
	binary.Read(r, binary.BigEndian, &load.Procs_zombie)

	return nil
}

func (load *CPULoad) Encode() (uint8, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(len(load.Items)))

	for _,item := range load.Items {
		binary.Write(buf, binary.BigEndian, item.Rate_user)
		binary.Write(buf, binary.BigEndian, item.Rate_sys)
		binary.Write(buf, binary.BigEndian, item.Rate_iowait)
		binary.Write(buf, binary.BigEndian, item.Rate_idle)
	}

	return SPC_CPULoad, buf
}

func (load *CPULoad) Decode(splen uint8, r io.Reader) error {
	var n uint8

	if splen < 1 { return fmt.Errorf("CPULoad.Decode: invalid subpacket size (%d)", splen) }
	binary.Read(r, binary.BigEndian, &n)
	if splen != 1 + n * 4 { return fmt.Errorf("CPULoad.Decode: invalid subpacket size (%d)", splen) }

	load.Items = make([]CPUItem, n)
	for i := 0; i < int(n); i ++ {
		binary.Read(r, binary.BigEndian, &load.Items[i].Rate_user)
		binary.Read(r, binary.BigEndian, &load.Items[i].Rate_sys)
		binary.Read(r, binary.BigEndian, &load.Items[i].Rate_iowait)
		binary.Read(r, binary.BigEndian, &load.Items[i].Rate_idle)
	}

	return nil
}

func (load *MemoryLoad) Encode() (uint8, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, load.free)
	binary.Write(buf, binary.BigEndian, load.buffers)
	binary.Write(buf, binary.BigEndian, load.cached)
	binary.Write(buf, binary.BigEndian, load.dirty)
	binary.Write(buf, binary.BigEndian, load.active)
	binary.Write(buf, binary.BigEndian, load.swaptotal)
	binary.Write(buf, binary.BigEndian, load.swapfree)
	binary.Write(buf, binary.BigEndian, load.swapcached)
	return SPC_MemoryLoad, buf
}

func (load *MemoryLoad) Decode(splen uint8, r io.Reader) (err error) {
	if splen != 32 { return fmt.Errorf("MemoryLoad.Decode: invalid subpacket size (%d)", splen) }

	binary.Read(r, binary.BigEndian, &load.free)
	binary.Read(r, binary.BigEndian, &load.buffers)
	binary.Read(r, binary.BigEndian, &load.cached)
	binary.Read(r, binary.BigEndian, &load.dirty)
	binary.Read(r, binary.BigEndian, &load.active)
	binary.Read(r, binary.BigEndian, &load.swaptotal)
	binary.Read(r, binary.BigEndian, &load.swapfree)
	binary.Read(r, binary.BigEndian, &load.swapcached)

	return nil
}

func (load *IOLoad) Encode() (uint8, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(len(load.Items)))

	for _,item := range load.Items {
		binary.Write(buf, binary.BigEndian, uint8(len(item.name)))
		buf.Write([]byte(item.name))
		binary.Write(buf, binary.BigEndian, item.tps_read)
		binary.Write(buf, binary.BigEndian, item.tps_written)
		binary.Write(buf, binary.BigEndian, item.kbytes_read)
		binary.Write(buf, binary.BigEndian, item.kbytes_written)
	}

	return SPC_IOLoad, buf
}

func (load *IOLoad) Decode(splen uint8, r io.Reader) (err error) {
	var i, num_items, namelen uint8

	err = binary.Read(r, binary.BigEndian, &num_items)
	if err != nil { return }
	load.Items = make([]DiskItem, num_items);

	for i = 0; i < num_items; i ++ {
		err = binary.Read(r, binary.BigEndian, &namelen)
		if err != nil { return }
		namebuf := make([]byte, namelen)
		_,err = io.ReadFull(r, namebuf)
		if err != nil { return }
		load.Items[i].name = string(namebuf)
		err = binary.Read(r, binary.BigEndian, &load.Items[i].tps_read)
		if err != nil { return }
		err = binary.Read(r, binary.BigEndian, &load.Items[i].tps_written)
		if err != nil { return }
		err = binary.Read(r, binary.BigEndian, &load.Items[i].kbytes_read)
		if err != nil { return }
		err = binary.Read(r, binary.BigEndian, &load.Items[i].kbytes_written)
		if err != nil { return }
	}

	return nil
}

func (load *NetworkLoad) Encode() (uint8, *bytes.Buffer) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.BigEndian, uint8(len(load.Items)))

	for _,item := range load.Items {
		binary.Write(buf, binary.BigEndian, uint8(len(item.name)))
		buf.Write([]byte(item.name))
		binary.Write(buf, binary.BigEndian, item.pkts_read)
		binary.Write(buf, binary.BigEndian, item.pkts_written)
		binary.Write(buf, binary.BigEndian, item.kbytes_read)
		binary.Write(buf, binary.BigEndian, item.kbytes_written)
	}

	return SPC_NetworkLoad, buf
}

func (load *NetworkLoad) Decode(splen uint8, r io.Reader) (err error) {
	var i, num_items, namelen uint8

	err = binary.Read(r, binary.BigEndian, &num_items)
	if err != nil { return }
	load.Items = make([]InterfaceItem, num_items);

	for i = 0; i < num_items; i ++ {
		err = binary.Read(r, binary.BigEndian, &namelen)
		if err != nil { return }
		namebuf := make([]byte, namelen)
		_,err = io.ReadFull(r, namebuf)
		if err != nil { return }
		load.Items[i].name = string(namebuf)
		err = binary.Read(r, binary.BigEndian, &load.Items[i].pkts_read)
		if err != nil { return }
		err = binary.Read(r, binary.BigEndian, &load.Items[i].pkts_written)
		if err != nil { return }
		err = binary.Read(r, binary.BigEndian, &load.Items[i].kbytes_read)
		if err != nil { return }
		err = binary.Read(r, binary.BigEndian, &load.Items[i].kbytes_written)
		if err != nil { return }
	}

	return nil
}

func EncodeSubpacket(sp Subpacket, w io.Writer) error {
	spcode,buf := sp.Encode()
	if buf.Len() > 255 { panic("Subpacket size overflow") }

	binary.Write(w, binary.BigEndian, spcode)
	binary.Write(w, binary.BigEndian, uint8(buf.Len()))
	buf.WriteTo(w)

	return nil
}

func (m *LoadMessage) Encode(w io.Writer) error {
	var err error

	// message header
	binary.Write(w, binary.BigEndian, uint8(MessageVersion))
	binary.Write(w, binary.BigEndian, m.Timestamp)
	binary.Write(w, binary.BigEndian, m.Interval)

	// load data
	if err = EncodeSubpacket(&m.Proc_load, w); err != nil { return err }
	if err = EncodeSubpacket(&m.Cpu_load, w); err != nil { return err }
	if err = EncodeSubpacket(&m.Mem_load, w); err != nil { return err }
	if err = EncodeSubpacket(&m.Io_load, w); err != nil { return err }
	if err = EncodeSubpacket(&m.Net_load, w); err != nil { return err }

	return nil
}

func (m *LoadMessage) Decode(r io.Reader) error {
	var n int
	var err error
	var version, spcode, splen uint8
	buf := make([]byte, 1)

	err = binary.Read(r, binary.BigEndian, &version)
	if err != nil { return err }
	if version != MessageVersion { return fmt.Errorf("version mismatch") }
	err = binary.Read(r, binary.BigEndian, &m.Timestamp)
	if err != nil { return err }
	err = binary.Read(r, binary.BigEndian, &m.Interval)
	if err != nil { return err }

	for n,err = r.Read(buf); n > 0; n,err = r.Read(buf) {
		spcode = buf[0]
		if n,err = r.Read(buf); n != 1 { return fmt.Errorf("premature subpacket") }
		splen = buf[0]
		spbuf := make([]byte, splen)
		n,err = io.ReadFull(r, spbuf)
		if err != nil { return fmt.Errorf("premature subpacket") }
		spreader := bytes.NewReader(spbuf)

		switch spcode {
		case SPC_ProcLoad:
			err = m.Proc_load.Decode(splen, spreader)
			if err != nil { return err }
		case SPC_CPULoad:
			err = m.Cpu_load.Decode(splen, spreader)
			if err != nil { return err }
		case SPC_MemoryLoad:
			err = m.Mem_load.Decode(splen, spreader)
			if err != nil { return err }
		case SPC_IOLoad:
			err = m.Io_load.Decode(splen, spreader)
			if err != nil { return err }
		case SPC_NetworkLoad:
			err = m.Net_load.Decode(splen, spreader)
			if err != nil { return err }
		default:
			return fmt.Errorf("unknown subpacket code")
		}
	}

	return nil
}
