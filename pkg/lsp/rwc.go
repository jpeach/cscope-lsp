package lsp

import (
	"io"
)

type rwc struct {
	write io.Writer
	read  io.Reader
}

func (r *rwc) Read(p []byte) (int, error) {
	return r.read.Read(p)
}

func (r *rwc) Write(p []byte) (int, error) {
	return r.write.Write(p)
}

func (r *rwc) Close() error {
	// Don't close, because really the Server owns these files.
	return nil
}

func (r *rwc) Tee(out io.Writer) io.ReadWriteCloser {
	return &rwc{
		write: io.MultiWriter(r.write, out),
		read:  io.TeeReader(r.read, out),
	}
}
