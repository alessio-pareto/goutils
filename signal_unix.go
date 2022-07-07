//go:build linux || darwin

package goutils

import (
	"os"
)

func sendCtrlC(p *os.Process) error {
	return p.Signal(os.Interrupt)
}

func restoreConsoleCtrlHandler() error {
	return nil
}