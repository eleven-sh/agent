package sshserver

import (
	"os/exec"
	"os/user"

	"github.com/eleven-sh/agent/config"
)

type sessionCmdBuilder struct {
	user *user.User
}

func newSessionCmdBuilder(
	user *user.User,
) *sessionCmdBuilder {

	return &sessionCmdBuilder{
		user: user,
	}
}

func (s *sessionCmdBuilder) build(args ...string) *exec.Cmd {
	cmdToBuildArgs := []string{
		"--set-home",
		"--login",
		"--user",
		s.user.Username,
	}

	return exec.Command("sudo", append(cmdToBuildArgs, args...)...)
}

func (s *sessionCmdBuilder) buildShell() *exec.Cmd {
	return s.build()
}

func (s *sessionCmdBuilder) buildShellPTY() *exec.Cmd {
	cmdToBuildArgs := []string{
		"login",
		"-f",
		s.user.Username,
	}

	return exec.Command("sudo", cmdToBuildArgs...)
}

func (s *sessionCmdBuilder) buildExec(cmd string) *exec.Cmd {
	return s.build(
		config.ElevenUserShellPath,
		"-l",
		"-c",
		cmd,
	)
}
