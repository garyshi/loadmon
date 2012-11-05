package sutils

import (
	"io"
	"bufio"
	"strconv"
)

func ReadLines(r io.Reader, f func(line string) (err error)) (err error) {
	var line string

	reader := bufio.NewReader(r)
	line, err = reader.ReadString('\n')
	for err == nil {
		err = f(line)
		if err != nil { return err }
		line, err = reader.ReadString('\n')
	}
	if err == io.EOF { err = nil }
	return
}

func Fields2Int(fields []string, cols []int) (rslt []int64, err error) {
	var r int64
	for i := range cols {
		r, err = strconv.ParseInt(fields[i], 0, 64)
		if err != nil { return }
		rslt = append(rslt, r)
	}
	return
}