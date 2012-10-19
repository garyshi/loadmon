package main

import (
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
}

func LogFileName(logdir, basename string, t *time.Time) (filename string) {
	if t == nil {
		filename = fmt.Sprintf("%s.log", basename)
	} else {
		filename = fmt.Sprintf("%s-%s.log", basename, t.Format("20060102"))
	}

	if len(logdir) > 0 {
		filename = fmt.Sprintf("%s/%s", logdir, filename)
	}

	return
}

func OpenLogFile(filename string, mode int) (logfile *LogFile, err error) {
	logfile = &LogFile{mode:mode}
	switch logfile.mode {
	case MODE_APPEND:
		mode = os.O_CREATE|os.O_WRONLY|os.O_APPEND
	case MODE_REWRITE:
		mode = os.O_CREATE|os.O_WRONLY|os.O_TRUNC
	case MODE_READ:
		mode = os.O_RDONLY
	default:
		return nil, fmt.Errorf("invalid mode")
	}

	logfile.file,err = os.OpenFile(filename, mode, 0644)
	if err != nil { return nil, err }

	return
}

// TODO: switch to new logfile as time goes on...
func (logfile *LogFile) WriteMessage(buf []byte) error {
	if logfile.mode != MODE_APPEND && logfile.mode != MODE_REWRITE {
		return fmt.Errorf("invalid mode")
	}

	ts := GetTimestamp()
	//logfile.file.Seek(0, os.SEEK_END)
	w := bufio.NewWriter(logfile.file)
	binary.Write(w, binary.BigEndian, ts)
	binary.Write(w, binary.BigEndian, uint16(len(buf)))
	w.Write(buf)
	return w.Flush()
}

func (logfile *LogFile) ReadMessage() (ts uint32, buf []byte, err error) {
	if logfile.mode != MODE_APPEND && logfile.mode != MODE_REWRITE {
		err = fmt.Errorf("invalid mode")
		return
	}

	return
}
