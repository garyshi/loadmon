package sutils

import (
	"io"
	"bufio"
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
