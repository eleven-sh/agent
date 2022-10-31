package sshserver

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"regexp"
	"strings"

	"github.com/creack/pty"
	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/env"
	"github.com/gliderlabs/ssh"
)

type sessionHandler struct {
	cmdBuilder *sessionCmdBuilder
}

func newSessionHandler(
	cmdBuilder *sessionCmdBuilder,
) *sessionHandler {

	return &sessionHandler{
		cmdBuilder: cmdBuilder,
	}
}

func (s *sessionHandler) handle(sshSession ssh.Session) {
	var sessionError error

	defer func() {
		if sessionError != nil {
			log.Printf(
				"[SSH server] Error during SSH session: %v",
				sessionError,
			)

			sshSession.Exit(1)
			return
		}

		sshSession.Exit(0)
	}()

	if len(sshSession.Command()) == 0 { // "shell" session
		_, _, hasPTY := sshSession.Pty()

		if hasPTY {
			sessionError = s.handleShellPTY(sshSession)
			return
		}

		sessionError = s.handleShell(sshSession)
		return
	}

	// "exec" session
	sessionError = s.handleExec(sshSession)
}

func (s *sessionHandler) handleShell(sshSession ssh.Session) error {
	_, _, hasPTY := sshSession.Pty()

	if hasPTY {
		return errors.New("expected no PTY, got PTY")
	}

	agentConfig, err := env.LoadConfigIfExists(
		config.ElevenAgentConfigFilePath,
	)

	if err != nil {
		return err
	}

	shellCmd := s.cmdBuilder.buildShell()

	shellCmdStdin, err := shellCmd.StdinPipe()

	if err != nil {
		return err
	}

	shellCmdStdout, err := shellCmd.StdoutPipe()

	if err != nil {
		return err
	}

	shellCmdStderr, err := shellCmd.StderrPipe()

	if err != nil {
		return err
	}

	stdinChan := make(chan error, 1)
	go func() {
		codeServerBinRegExp := regexp.MustCompile(`code-server`)
		codeServerBinRegExpMatched := false

		codeServerStartFlagRegExp := regexp.MustCompile(`--start-server`)

		stdin := bufio.NewReader(sshSession)

		for {
			shellLine, err := stdin.ReadString('\n')

			if err != nil && errors.Is(err, io.EOF) {
				stdinChan <- nil
				break
			}

			if err != nil {
				stdinChan <- err
				break
			}

			vscExtensions := []string{}

			if agentConfig != nil {
				vscExtensions = agentConfig.Workspace.VSCode.Extensions
			}

			if len(vscExtensions) > 0 {

				if codeServerBinRegExp.MatchString(shellLine) {
					codeServerBinRegExpMatched = true
				}

				if codeServerBinRegExpMatched &&
					codeServerStartFlagRegExp.MatchString(shellLine) {

					shellLine = codeServerStartFlagRegExp.ReplaceAllString(
						shellLine,
						"--start-server --install-extension "+
							strings.Join(vscExtensions, " --install-extension "),
					)
				}
			}

			_, err = shellCmdStdin.Write([]byte(shellLine))

			if err != nil {
				stdinChan <- err
				break
			}
		}
	}()

	stdoutChan := make(chan error, 1)
	go func() {
		_, err := io.Copy(sshSession, shellCmdStdout)
		stdoutChan <- err
	}()

	stderrChan := make(chan error, 1)
	go func() {
		_, err := io.Copy(sshSession, shellCmdStderr)
		stderrChan <- err
	}()

	if err := shellCmd.Start(); err != nil {
		return err
	}

	stdoutErr := <-stdoutChan
	stderrErr := <-stderrChan

	if stdoutErr != nil {
		return stdoutErr
	}

	if stderrErr != nil {
		return stderrErr
	}

	// It is incorrect to call "Wait()"
	// before all reads from the pipes have completed.
	// See "StderrPipe()" / "StdoutPipe()" documentation.
	return shellCmd.Wait()
}

func (s *sessionHandler) handleShellPTY(sshSession ssh.Session) error {
	ptyReq, windowChan, hasPTY := sshSession.Pty()

	if !hasPTY {
		return errors.New("expected PTY, got no PTY")
	}

	shellCmd := s.cmdBuilder.buildShellPTY()

	shellCmd.Env = append(
		shellCmd.Env,
		fmt.Sprintf("TERM=%s", ptyReq.Term),
	)

	shellCmdPty, err := pty.Start(shellCmd)

	if err != nil {
		return err
	}

	go func() {
		for window := range windowChan {
			setWindowSize(
				shellCmdPty,
				window.Width,
				window.Height,
			)
		}
	}()

	go func() {
		io.Copy(shellCmdPty, sshSession) // stdin
	}()

	io.Copy(sshSession, shellCmdPty) // stdout

	return shellCmd.Wait()
}

func (s *sessionHandler) handleExec(sshSession ssh.Session) error {
	passedCmd := sshSession.RawCommand()

	if len(passedCmd) == 0 {
		return errors.New("expected command, got nothing")
	}

	cmdToExec := s.cmdBuilder.buildExec(
		passedCmd,
	)

	cmdToExec.Stdin = sshSession
	cmdToExec.Stdout = sshSession
	cmdToExec.Stderr = sshSession

	if err := cmdToExec.Start(); err != nil {
		return err
	}

	// We use ".Process.Wait()" here given that
	// ".Wait()" will wait indefinitly for "Stdin"
	// (the SSH channel) to close before returning.
	cmdState, err := cmdToExec.Process.Wait()

	if err != nil {
		return err
	}

	if cmdExitCode := cmdState.ExitCode(); cmdExitCode != 0 {
		return fmt.Errorf(
			"the command \"%s\" has returned a non-zero (%d) exit code",
			passedCmd,
			cmdExitCode,
		)
	}

	return nil
}
