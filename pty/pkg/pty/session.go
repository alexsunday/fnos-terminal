package pty

import (
	"io"
)

type Session interface {
	io.ReadWriteCloser
	Resize(cols, rows uint16) error
}

type Winsize struct {
	Cols uint16 `json:"cols"`
	Rows uint16 `json:"rows"`
}
