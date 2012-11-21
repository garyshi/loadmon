package main

import (
	"io"
	"os"
	"fmt"
	"time"
	"bufio"
	"encoding/binary"
)

const (
	MODE_READ = iota
	MODE_APPEND
	MODE_REWRITE
)

type LogFile struct {
	mode int
	file *os.File
	filename, basename string
	basedate int
}

func OpenRotateLogFile(basename string, t *time.Time, mode int) (logfile *LogFile, err error) {
	logfile = &LogFile{mode:mode, basename:basename}
	logfile.filename = fmt.Sprintf("%s-%s.log", basename, t.Format("20060102"))
	logfile.basedate = int(ToTimestamp(*t) / 86400)
	return logfile, logfile.Open()
}

func OpenLogFile(filename string, mode int) (logfile *LogFile, err error) {
	logfile = &LogFile{mode:mode, filename:filename}
	return logfile, logfile.Open()
}

func (logfile *LogFile) Open() (err error) {
	var mode int

	if logfile.file != nil { return fmt.Errorf("logfile already open") }

	switch logfile.mode {
	case MODE_APPEND:
		mode = os.O_CREATE|os.O_WRONLY|os.O_APPEND
	case MODE_REWRITE:
		mode = os.O_CREATE|os.O_WRONLY|os.O_TRUNC
	case MODE_READ:
		mode = os.O_RDONLY
	default:
		return fmt.Errorf("invalid mode")
	}

	logfile.file,err = os.OpenFile(logfile.filename, mode, 0644)

	return
}

func (logfile *LogFile) Close() {
	if logfile.file != nil {
		logfile.file.Close()
		logfile.file = nil
	}
}

func (logfile *LogFile) WriteMessage(buf []byte) error {
	if logfile.mode != MODE_APPEND && logfile.mode != MODE_REWRITE {
		return fmt.Errorf("invalid mode")
	}

	now := time.Now()
	ts := ToTimestamp(now)
	if logfile.basedate != 0 {
		basedate := int(ts / 86400)
		if basedate != logfile.basedate {
			logfile.file.Close()
			logfile.file = nil
			logfile.filename = fmt.Sprintf("%s-%s.log", logfile.basename, now.Format("20060102"))
			logfile.basedate = basedate
			err := logfile.Open()
			if err != nil { return err }
		}
	}

	//logfile.file.Seek(0, os.SEEK_END)
	w := bufio.NewWriter(logfile.file)
	binary.Write(w, binary.BigEndian, ts)
	binary.Write(w, binary.BigEndian, uint16(len(buf)))
	w.Write(buf)
	return w.Flush()
}

func (logfile *LogFile) ReadMessage() (ts uint32, buf []byte, err error) {
	if logfile.mode != MODE_READ {
		err = fmt.Errorf("invalid mode")
		return
	}

	var msglen uint16
	err = binary.Read(logfile.file, binary.BigEndian, &ts)
	if err != nil { return }
	err = binary.Read(logfile.file, binary.BigEndian, &msglen)
	if err != nil { return }
	buf = make([]byte, msglen)
	_,err = io.ReadFull(logfile.file, buf)

	return
}
