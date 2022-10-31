package state

import (
	"bytes"
	"log"
	"os/exec"
	"sync"
	"syscall"
	"time"

	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/internal/network"
	"github.com/prometheus/procfs"
)

type process struct {
	cmdWD     env.ConfigLongRunningProcessWD
	cmdString env.ConfigLongRunningProcessCmd
	cmd       *exec.Cmd
	doneChan  chan struct{}
}

var currentProcesses = map[env.ConfigLongRunningProcessWD]*process{}
var currentProcessesLock sync.Mutex

func ReconcileLongRunningProcesses(
	newProcesses env.ConfigLongRunningProcesses,
) error {

	currentProcessesLock.Lock()
	defer currentProcessesLock.Unlock()

	for currentProcessWD, currentProcess := range currentProcesses {

		newProcessCmd, newProcessExistsInWD := newProcesses[currentProcessWD]

		if newProcessExistsInWD && newProcessCmd == currentProcess.cmdString {
			continue
		}

		clearProcess(currentProcess)
	}

	for newProcessWD, newProcessCmd := range newProcesses {

		if _, alreadyRun := currentProcesses[newProcessWD]; alreadyRun {
			continue
		}

		processToStart := newProcess(
			newProcessWD,
			newProcessCmd,
			nil,
		)

		cmd, err := startProcess(processToStart)

		if err != nil {
			log.Printf(
				"[Forever] Error when starting process %s:%s: %v",
				newProcessWD,
				newProcessCmd,
				err,
			)

			continue
		}

		processToStart.cmd = cmd

		currentProcesses[newProcessWD] = processToStart

		go waitForProcess(processToStart)
	}

	return nil
}

func newProcess(
	cmdWD env.ConfigLongRunningProcessWD,
	cmdString env.ConfigLongRunningProcessCmd,
	cmd *exec.Cmd,
) *process {

	return &process{
		cmdWD:     cmdWD,
		cmdString: cmdString,
		cmd:       cmd,
		doneChan:  make(chan struct{}),
	}
}

func buildProcessCmd(cmdWD, cmdString string) *exec.Cmd {
	cmd := exec.Command(
		config.ElevenUserShellPath,
		"-i",
		"-c",
		cmdString,
	)

	cmd.Dir = cmdWD
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}

	return cmd
}

func startProcess(p *process) (*exec.Cmd, error) {
	cmd := buildProcessCmd(
		string(p.cmdWD),
		string(p.cmdString),
	)

	return cmd, cmd.Start()
}

// Killing a child process and all of its children in Go
// See: https://stackoverflow.com/questions/22470193/why-wont-go-kill-a-child-process-correctly
// and https://medium.com/@felixge/killing-a-child-process-and-all-of-its-children-in-go-54079af94773
func killProcess(cmd *exec.Cmd) error {
	pgid, err := syscall.Getpgid(cmd.Process.Pid)

	if err != nil {
		return err
	}

	return syscall.Kill(-pgid, syscall.SIGINT)
}

func waitForProcess(p *process) error {
	unexpectedProcessExit := false

	go func() {
		<-p.doneChan

		if unexpectedProcessExit {
			return
		}

		if err := killProcess(p.cmd); err != nil {
			log.Printf(
				"[Forever] Error when killing process %s:%s: %v",
				p.cmdWD,
				p.cmdString,
				err,
			)
		}
	}()

	err := p.cmd.Wait()

	select {
	case <-p.doneChan:
		return nil
	default:
		// Needs to be set BEFORE "clearProcess",
		// otherwise the goroutine may try to kill
		// already killed process
		unexpectedProcessExit = true
		clearProcess(p)

		log.Printf(
			"[Forever] Unexpected exit for process %s:%s: %v",
			p.cmdWD,
			p.cmd,
			err,
		)

		return err
	}
}

func clearProcess(p *process) {
	close(p.doneChan)
	delete(currentProcesses, p.cmdWD)
}

func StartProcessAndWaitForSleep(
	cmdWD string,
	cmdString string,
	heartbeatChan <-chan error,
) (exitOutput string, exitErrMsg string, returnedError error) {

	var cmdStdBuf bytes.Buffer
	cmd := buildProcessCmd(cmdWD, cmdString)
	cmd.Stdout = &cmdStdBuf
	cmd.Stderr = &cmdStdBuf

	cmdProcess := newProcess(
		env.ConfigLongRunningProcessWD(cmdWD),
		env.ConfigLongRunningProcessCmd(cmdString),
		cmd,
	)

	initialTCPConnInodes := map[uint64]bool{}
	initialTCPConns, err := network.GetOpenedTCPConns()

	if err != nil {
		returnedError = err
		return
	}

	for _, tcpConn := range initialTCPConns {
		initialTCPConnInodes[tcpConn.Inode] = true
	}

	if err := cmd.Start(); err != nil {
		returnedError = err
		return
	}

	cmdExited := false
	cmdExitedChan := make(chan error, 1)
	go func() {
		cmdExitedChan <- waitForProcess(cmdProcess)
		cmdExited = true
	}()

	cmdStartedChan := make(chan error, 1)
	go func() {
		cmdProcSleepSinceSeconds := 0
		cmdProcOpenedTCPConn := false

		cmdStdBufLen := cmdStdBuf.Len()

		processGrpID, err := syscall.Getpgid(cmd.Process.Pid)

		if err != nil {
			cmdStartedChan <- err
			return
		}

		for {
			if cmdExited {
				return
			}

			processes, err := procfs.AllProcs()

			if err != nil {
				cmdStartedChan <- err
				return
			}

			processAndChildSleep := true

			for _, process := range processes {
				st, err := process.Stat()

				if err != nil {
					// Race condition
					continue
				}

				if process.PID != cmd.Process.Pid && st.PGRP != processGrpID {
					continue
				}

				if st.State != "S" {
					processAndChildSleep = false
					break
				}
			}

			if !cmdProcOpenedTCPConn {
				openedTCPConns, err := network.GetOpenedTCPConns()

				if err != nil {
					cmdStartedChan <- err
					return
				}

				for _, conn := range openedTCPConns {
					if conn.St != uint64(network.TCPConnStatusListening) {
						continue
					}

					if _, connExisted := initialTCPConnInodes[conn.Inode]; connExisted {
						continue
					}

					cmdProcOpenedTCPConn = true
					break
				}
			}

			if processAndChildSleep {
				cmdProcSleepSinceSeconds++
			}

			if !processAndChildSleep {
				cmdProcSleepSinceSeconds = 0
			}

			if cmdStdBufLen != cmdStdBuf.Len() {
				cmdStdBufLen = cmdStdBuf.Len()
				cmdProcSleepSinceSeconds = 0
			}

			// cmdProcSleepSinceSeconds is not used currently.
			// Maybe in the future, following user feedback.
			if cmdProcSleepSinceSeconds >= 0 && cmdProcOpenedTCPConn {
				close(cmdStartedChan)
				return
			}

			time.Sleep(1 * time.Second)
		}
	}()

	select {
	case err := <-cmdExitedChan:
		errMessage := ""
		if err != nil {
			errMessage = " (" + err.Error() + ")"
		}

		exitOutput = cmdStdBuf.String()
		exitErrMsg = "unexpected command exit" + errMessage
		return
	case err := <-cmdStartedChan:
		if err != nil {
			clearProcess(cmdProcess)

			returnedError = err
			return
		}

		returnedError = saveLongRunningProcess(cmdProcess)
		return
	case err := <-heartbeatChan:
		clearProcess(cmdProcess)

		returnedError = err
		return
	}
}

func saveLongRunningProcess(p *process) error {
	agentConfig, err := env.LoadConfig(
		config.ElevenAgentConfigFilePath,
	)

	if err != nil {
		return err
	}

	agentConfig.LongRunningProcesses[p.cmdWD] = p.cmdString

	currentProcessesLock.Lock()
	defer currentProcessesLock.Unlock()

	err = env.SaveConfigAsFile(
		config.ElevenAgentConfigFilePath,
		agentConfig,
	)

	if err != nil {
		return err
	}

	currentProcesses[p.cmdWD] = p

	return nil
}
