// +build !linux

package lsp

import "syscall"

func procattr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{}
}
