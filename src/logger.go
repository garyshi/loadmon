package main

import (
	"io"
)

func (m *LoadMessage) WriteLog(w io.Writer) error {
	return nil
}

func (m *LoadMessage) ReadLog(r io.Reader) error {
	return nil
}
