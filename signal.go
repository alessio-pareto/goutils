package goutils

import (
	"os"
)

// Sends CRTL-C signal to the process.
// On Linux and Darwin has default behaviour
// On Windows first suppresses the handler of the CTRL-C signal of the calling
// program, then sends the signal to all the processes belonging to the same group.
// After all the programs have exited, only on Windows, you should call the
// function RestoreConsoleCtrlHandler to enable the default behaviour of the calling
// program
func SendCtrlC(p *os.Process) error {
	return sendCtrlC(p)
}

// Restores the default behaviour of the CTRL-C signal on Windows, otherwise does nothing
func RestoreConsoleCtrlHandler() error {
	return restoreConsoleCtrlHandler()
}