package goutils

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
)

// Program wraps the standard exec.Cmd structure provided by Go
// but implements an easy way to send a real CTRL-C to the child
// process (which can be handled it), also on Windows (even though its not possible
// using the standard libraries). However on windows if you create more than one
// child process and you call Stop() on one of the programs, all the childs
// will receive the signal, even if they are created using the standard libraries
type Program struct {
	dir string
	execName string
	args  []string
	exitC 	chan struct{}
	waiting int
	exec    *exec.Cmd
	redirect bool
	lastExitCode int
}

// Creates a new program. You can specify the working directory, or leave it
// blank to use the current one, and if the standard output and error should be redirected.
// Currently standard input redirect its not supported but will be implemented soon
func NewProgram(dir string, redirect bool, execName string, args ...string) (*Program, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("exec: directory not found")
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("exec: dir is not a directory")
	}

	p := &Program{
		dir:      dir,
		execName: execName,
		args:     args,
		exitC:    make(chan struct{}, 3),
		redirect: redirect,
		lastExitCode: -1,
	}

	return p, nil
}

// Tells whether the program is running or not
func (p *Program) IsRunning() bool {
	return p.exec != nil
}

// Starts the program if its not already running, otherwise returns an error
func (p *Program) Start() error {
	if p.IsRunning() {
		return fmt.Errorf("program %s already running", p.execName)
	}

	err := p.start()
	if err != nil {
		return fmt.Errorf("error starting program %s: %v", p.execName, err)
	}

	return nil
}

// Reports the last exit code of the program. If it returns -1, it means
// that the program never started or has not exited yet
func (p *Program) LastExitCode() int {
	return p.lastExitCode
}

func (p *Program) start() error {
	p.exec = exec.Command(p.execName, p.args...)
	if p.dir != "" {
		p.exec.Dir = p.dir
	}

	if p.redirect {
		p.exec.Stdout = os.Stdout
		p.exec.Stderr = os.Stderr
	}

	err := p.exec.Start()
	if err != nil {
		p.exec = nil
		return err
	}

	go p.wait()
	return nil
}

func (p *Program) wait() {
	if p.exec == nil {
		return
	}

	p.exec.Wait()
	
	p.lastExitCode = p.exec.ProcessState.ExitCode()
	p.exec = nil

	for p.waiting > 0 {
		p.exitC <- struct{}{}
		p.waiting --
	}
}

// Waits the program until its exits, or returns instantly
// if the program isn't running
func (p *Program) Wait() {
	if p.exec == nil {
		return
	}

	p.waiting ++
	<- p.exitC
}

// Starts the program and waits for its exit
func (p *Program) Run() error {
	err := p.Start()
	if err != nil {
		return err
	}

	p.Wait()
	return nil
}

// Forces the program to exit
func (p *Program) Kill() {
	if p.exec != nil {
		p.exec.Process.Kill()
		p.Wait()
	}
}

// Sends a CTRL-C signal to the process. On Windows this signal is sent
// to all existing child processes spawned from every library
func (p *Program) Stop() error {
	if p.exec == nil {
		return nil
	}

	wg := new(sync.WaitGroup)
	wg.Add(1)

	go func() {
		p.Wait()
		wg.Done()
	}()

	err := SendCtrlC(p.exec.Process)
	if err != nil {
		panic(err)
	}
	wg.Wait()

	err = RestoreConsoleCtrlHandler()
	if err != nil {
		panic(err)
	}

	return nil
}

func (p *Program) String() string {
	var state string
	if p.IsRunning() {
		state = fmt.Sprintf("Running - %d", p.exec.Process.Pid)
	} else {
		state = "Stopped"
	}
	return fmt.Sprintf("%s (%s)", p.execName, state)
}
