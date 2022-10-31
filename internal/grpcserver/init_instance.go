package grpcserver

import (
	"bufio"
	_ "embed"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"

	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/proto"
)

//go:embed init_instance.sh
var initInstanceScript string

func (*agentServer) InitInstance(
	req *proto.InitInstanceRequest,
	stream proto.Agent_InitInstanceServer,
) error {

	initInstanceScriptFilePath, err := createInitInstanceScriptFile()

	if err != nil {
		return err
	}

	defer os.Remove(initInstanceScriptFilePath)

	initInstanceCmd := buildInitInstanceCmd(initInstanceScriptFilePath, req)

	stdoutReader, err := buildInitInstanceCmdStdoutReader(initInstanceCmd)

	if err != nil {
		return err
	}

	stderrReader, err := buildInitInstanceCmdStderrReader(initInstanceCmd)

	if err != nil {
		return err
	}

	if err := initInstanceCmd.Start(); err != nil {
		return err
	}

	stdoutHandlerChan := make(chan error, 1)

	go func() {
		stdoutHandlerChan <- handleInitInstanceCmdOutput(
			stdoutReader,
			stream,
		)
	}()

	stderrHandlerChan := make(chan error, 1)

	go func() {
		stderrHandlerChan <- handleInitInstanceCmdOutput(
			stderrReader,
			stream,
		)
	}()

	stdoutHandlerErr := <-stdoutHandlerChan
	stderrHandlerErr := <-stderrHandlerChan

	if stdoutHandlerErr != nil {
		return stdoutHandlerErr
	}

	if stderrHandlerErr != nil {
		return stderrHandlerErr
	}

	// It is incorrect to call "Wait()"
	// before all reads from the pipes have completed.
	// See "StderrPipe()" / "StdoutPipe()" documentation.
	if err := initInstanceCmd.Wait(); err != nil {
		return fmt.Errorf(
			"unexpected init instance script exit (%v)",
			err,
		)
	}

	githubSSHPublicKeyContent, err := readGitHubSSHPublicKey(
		config.GitHubPublicSSHKeyFilePath,
	)

	if err != nil {
		return err
	}

	err = stream.Send(&proto.InitInstanceReply{
		GithubSshPublicKeyContent: &githubSSHPublicKeyContent,
	})

	if err != nil {
		return err
	}

	agentConfig := env.NewConfig()

	return env.PrepareWorkspace(
		agentConfig,
		req.EnvName,
		req.EnvRepos,
	)
}

func createInitInstanceScriptFile() (string, error) {
	initInstanceScriptFile, err := os.CreateTemp("", "eleven_init_script_*")

	if err != nil {
		return "", err
	}

	err = fillInitInstanceScriptFile(initInstanceScriptFile)

	if err != nil {
		return "", err
	}

	// Opened file cannot be executed at the same time.
	// Prevent "fork/exec text file busy" error.
	err = closeInitInstanceScriptFile(initInstanceScriptFile)

	if err != nil {
		return "", err
	}

	err = addExecPermsToInitInstanceScriptFile(initInstanceScriptFile)

	if err != nil {
		return "", err
	}

	return initInstanceScriptFile.Name(), nil
}

func fillInitInstanceScriptFile(initInstanceScriptFile *os.File) error {
	_, err := initInstanceScriptFile.WriteString(initInstanceScript)
	return err
}

func closeInitInstanceScriptFile(initInstanceScriptFile *os.File) error {
	return initInstanceScriptFile.Close()
}

func addExecPermsToInitInstanceScriptFile(initInstanceScriptFile *os.File) error {
	return os.Chmod(
		initInstanceScriptFile.Name(),
		os.FileMode(0700),
	)
}

func buildInitInstanceCmd(
	initInstanceScriptFilePath string,
	req *proto.InitInstanceRequest,
) *exec.Cmd {

	initInstanceCmd := exec.Command(initInstanceScriptFilePath)

	initInstanceCmd.Dir = path.Dir(initInstanceScriptFilePath)
	initInstanceCmd.Env = buildInitInstanceCmdEnvVars(req)

	return initInstanceCmd
}

func buildInitInstanceCmdEnvVars(req *proto.InitInstanceRequest) []string {
	return append(os.Environ(), []string{
		fmt.Sprintf("ELEVEN_CONFIG_DIR_PATH=%s", config.ElevenConfigDirPath),
		fmt.Sprintf("VSCODE_CONFIG_DIR_PATH=%s", config.VSCodeConfigDirPath),
		fmt.Sprintf("ENV_NAME_SLUG=%s", req.EnvNameSlug),
		fmt.Sprintf("GITHUB_USER_EMAIL=%s", req.GithubUserEmail),
		fmt.Sprintf("USER_FULL_NAME=%s", req.UserFullName),
	}...)
}

func buildInitInstanceCmdStderrReader(initInstanceCmd *exec.Cmd) (*bufio.Reader, error) {
	stderrPipe, err := initInstanceCmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	return bufio.NewReader(stderrPipe), nil
}

func buildInitInstanceCmdStdoutReader(initInstanceCmd *exec.Cmd) (*bufio.Reader, error) {
	stdoutPipe, err := initInstanceCmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	return bufio.NewReader(stdoutPipe), nil
}

func handleInitInstanceCmdOutput(
	outputReader *bufio.Reader,
	stream proto.Agent_InitInstanceServer,
) error {

	for {
		outputLine, err := outputReader.ReadString('\n')

		if err != nil && errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		err = stream.Send(&proto.InitInstanceReply{
			LogLine: outputLine,
		})

		if err != nil {
			return err
		}
	}

	return nil
}

func readGitHubSSHPublicKey(sshPublicKeyFilePath string) (string, error) {
	sshPublicKeyContent, err := os.ReadFile(sshPublicKeyFilePath)

	if err != nil {
		return "", err
	}

	return string(sshPublicKeyContent), nil
}
