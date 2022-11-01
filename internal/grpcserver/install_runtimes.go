package grpcserver

import (
	"bufio"
	"embed"
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

//go:embed runtimes/*
var runtimeScripts embed.FS

func (*agentServer) InstallRuntimes(
	req *proto.InstallRuntimesRequest,
	stream proto.Agent_InstallRuntimesServer,
) error {

	type runtime struct {
		name    string
		version string
	}
	orderedRuntimes := []runtime{}
	var rubyRuntime *runtime

	for runtimeName, runtimeVersion := range req.Runtimes {
		if runtimeName == "ruby" {
			rubyRuntime = &runtime{
				name:    runtimeName,
				version: runtimeVersion,
			}
			continue
		}

		orderedRuntimes = append(orderedRuntimes, runtime{
			name:    runtimeName,
			version: runtimeVersion,
		})
	}

	// RVM needs to be the last change to the $PATH env var
	if rubyRuntime != nil {
		orderedRuntimes = append(orderedRuntimes, *rubyRuntime)
	}

	for _, runtime := range orderedRuntimes {
		err := stream.Send(&proto.InstallRuntimesReply{
			LogLineHeader: fmt.Sprintf(
				"Installing %s@%s",
				runtime.name,
				runtime.version,
			),
		})

		if err != nil {
			return err
		}

		installRuntimeScriptFilePath, err := createInstallRuntimeScriptFile(
			runtime.name,
		)

		if err != nil {
			return err
		}

		defer os.Remove(installRuntimeScriptFilePath)

		installRuntimeCmd := buildInstallRuntimeCmd(
			installRuntimeScriptFilePath,
			runtime.version,
		)

		stdoutReader, err := buildInstallRuntimeCmdStdoutReader(installRuntimeCmd)

		if err != nil {
			return err
		}

		stderrReader, err := buildInstallRuntimeCmdStderrReader(installRuntimeCmd)

		if err != nil {
			return err
		}

		if err := installRuntimeCmd.Start(); err != nil {
			return err
		}

		stdoutHandlerChan := make(chan error, 1)

		go func() {
			stdoutHandlerChan <- handleInstallRuntimeCmdOutput(
				stdoutReader,
				stream,
			)
		}()

		stderrHandlerChan := make(chan error, 1)

		go func() {
			stderrHandlerChan <- handleInstallRuntimeCmdOutput(
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
		if err := installRuntimeCmd.Wait(); err != nil {
			return fmt.Errorf(
				"error installing %s@%s (%v)",
				runtime.name,
				runtime.version,
				err,
			)
		}
	}

	agentConfig, err := env.LoadConfig(
		config.ElevenAgentConfigFilePath,
	)

	if err != nil {
		return err
	}

	agentConfig.Workspace.SetRuntimes(req.Runtimes)

	return env.SaveConfigAsFile(
		config.ElevenAgentConfigFilePath,
		agentConfig,
	)
}

func createInstallRuntimeScriptFile(runtime string) (string, error) {
	installRuntimeScriptFile, err := os.CreateTemp("", "eleven_install_runtime_*")

	if err != nil {
		return "", err
	}

	err = fillInstallRuntimeScriptFile(
		runtime,
		installRuntimeScriptFile,
	)

	if err != nil {
		return "", err
	}

	// Opened file cannot be executed at the same time.
	// Prevent "fork/exec text file busy" error.
	err = closeInstallRuntimeScriptFile(installRuntimeScriptFile)

	if err != nil {
		return "", err
	}

	err = addExecPermsToInstallRuntimeScriptFile(installRuntimeScriptFile)

	if err != nil {
		return "", err
	}

	return installRuntimeScriptFile.Name(), nil
}

func fillInstallRuntimeScriptFile(
	runtime string,
	installRuntimeScriptFile *os.File,
) error {

	installRuntimeScript, err := runtimeScripts.ReadFile("runtimes/" + runtime + ".sh")

	if err != nil {
		return err
	}

	_, err = installRuntimeScriptFile.Write(installRuntimeScript)
	return err
}

func closeInstallRuntimeScriptFile(installRuntimeScriptFile *os.File) error {
	return installRuntimeScriptFile.Close()
}

func addExecPermsToInstallRuntimeScriptFile(installRuntimeScriptFile *os.File) error {
	return os.Chmod(
		installRuntimeScriptFile.Name(),
		os.FileMode(0700),
	)
}

func buildInstallRuntimeCmd(
	installRuntimeScriptFilePath string,
	runtimeVersion string,
) *exec.Cmd {

	installRuntimeCmd := exec.Command(installRuntimeScriptFilePath)

	installRuntimeCmd.Dir = path.Dir(installRuntimeScriptFilePath)
	installRuntimeCmd.Env = buildInstallRuntimeCmdEnvVars(runtimeVersion)

	return installRuntimeCmd
}

func buildInstallRuntimeCmdEnvVars(runtimeVersion string) []string {
	return append(os.Environ(), []string{
		// "SHELL" is used by installer like "nvm"
		// to automatically add required lines
		// to corresponding shell config file (".zshrc", ".bashrc", etc).
		//
		// "os.Environ()" contains an old value given that
		// the default shell for the user eleven is
		// modified in the "init_instance.sh" script.
		fmt.Sprintf("SHELL=%s", config.ElevenUserShellPath),
		fmt.Sprintf("RUNTIME_VERSION=%s", runtimeVersion),
	}...)
}

func buildInstallRuntimeCmdStderrReader(installRuntimeCmd *exec.Cmd) (*bufio.Reader, error) {
	stderrPipe, err := installRuntimeCmd.StderrPipe()

	if err != nil {
		return nil, err
	}

	return bufio.NewReader(stderrPipe), nil
}

func buildInstallRuntimeCmdStdoutReader(installRuntimeCmd *exec.Cmd) (*bufio.Reader, error) {
	stdoutPipe, err := installRuntimeCmd.StdoutPipe()

	if err != nil {
		return nil, err
	}

	return bufio.NewReader(stdoutPipe), nil
}

func handleInstallRuntimeCmdOutput(
	outputReader *bufio.Reader,
	stream proto.Agent_InstallRuntimesServer,
) error {

	for {
		outputLine, err := outputReader.ReadString('\n')

		if err != nil && errors.Is(err, io.EOF) {
			break
		}

		if err != nil {
			return err
		}

		err = stream.Send(&proto.InstallRuntimesReply{
			LogLine: outputLine,
		})

		if err != nil {
			return err
		}
	}

	return nil
}
