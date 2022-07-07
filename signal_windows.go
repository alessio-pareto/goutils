package goutils

import (
	"fmt"
	"os"
	"syscall"
)

func sendCtrlC(p *os.Process) error {
	d, e := loadKernel32()
	if e != nil {
		return e
	}

	if e = removeConsoleCtrlHandler(d); e != nil {
		return e
	}

	if e = generateConsoleCtrlEvent(d, p.Pid); e != nil {
		return e
	}

	return nil
}

func loadKernel32() (*syscall.DLL, error) {
	d, e := syscall.LoadDLL("kernel32.dll")
	if e != nil {
		return nil, fmt.Errorf("loadDLL: %v", e)
	}

	return d, e
}

func setConsoleCtrlHandler(d *syscall.DLL, flag bool) error {
	p, e := d.FindProc("SetConsoleCtrlHandler")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	a := 0
	if flag {
		a = 1
	}

	r, _, e := p.Call(0, uintptr(a))
	if r == 0 {
		return fmt.Errorf("setConsoleCtrlHandler: %v", e)
	}

	return nil
}

func removeConsoleCtrlHandler(d *syscall.DLL) error {
	return setConsoleCtrlHandler(d, true)
}

func restoreConsoleCtrlHandler() error {
	d, e := loadKernel32()
	if e != nil {
		return e
	}

	return setConsoleCtrlHandler(d, false)
}

func generateConsoleCtrlEvent(d *syscall.DLL, pid int) error {
	p, e := d.FindProc("GenerateConsoleCtrlEvent")
	if e != nil {
		return fmt.Errorf("findProc: %v", e)
	}

	r, _, e := p.Call(syscall.CTRL_C_EVENT, uintptr(pid))
	if r == 0 {
		return fmt.Errorf("generateConsoleCtrlEvent: %v", e)
	}

	return nil
}
