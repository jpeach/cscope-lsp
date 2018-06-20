// +build linux

package lsp

import (
	"os"
	"syscall"
)

func procattr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		Pdeathsig: os.SIGKILL,
	}

}
