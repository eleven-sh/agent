package forever

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/briandowns/spinner"
	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/proto"
	"github.com/jwalton/gchalk"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type Action string

const (
	ActionStop Action = "stop"
)

func Run(args []string) {
	handleError := func(errMessage string) {
		fmt.Println("Forever: Error: " + errMessage)
		os.Exit(1)
	}

	if len(args) == 0 {
		fmt.Println("Forever: Usage: \"forever {<command>|stop}\"")
		return
	}

	cmdWD, err := os.Getwd()

	if err != nil {
		handleError(err.Error())
		return
	}

	action := Action(args[0])

	if action == ActionStop {
		err := runStopAction(cmdWD)

		if err != nil {
			handleError(err.Error())
			return
		}

		fmt.Println("Forever: command stopped")
		return
	}

	err = runStartAction(cmdWD, strings.Join(args, " "))

	if err != nil {
		handleError(err.Error())
		return
	}

	fmt.Println("Forever: command started. Run \"forever stop\" in current path to stop.")
}

func runStartAction(cmdWD, cmd string) error {
	agentConfig, err := env.LoadConfig(
		config.ElevenAgentConfigFilePath,
	)

	if err != nil {
		return err
	}

	if cmd, processExist := agentConfig.LongRunningProcesses[env.ConfigLongRunningProcessWD(cmdWD)]; processExist {
		return fmt.Errorf(
			"\"%s\" is already running in current path. Run \"forever stop\" first%s",
			cmd,
			".", // bypass static-check linter
		)
	}

	spin := spinner.New(spinner.CharSets[26], 400*time.Millisecond)
	spin.Prefix = gchalk.Bold("Forever: waiting for command to listen on a port")

	spin.Start()
	reply, err := tryToStartLongRunningProcess(cmdWD, cmd)
	spin.Stop()

	if err != nil {
		return err
	}

	if len(reply.ErrorMessage) > 0 {

		if len(reply.ErrorOutput) > 0 {
			fmt.Println(reply.ErrorOutput)
		}

		return fmt.Errorf(reply.ErrorMessage)
	}

	return nil
}

func runStopAction(cmdWD string) error {
	agentConfig, err := env.LoadConfig(
		config.ElevenAgentConfigFilePath,
	)

	if err != nil {
		return err
	}

	if _, processExist := agentConfig.LongRunningProcesses[env.ConfigLongRunningProcessWD(cmdWD)]; !processExist {
		return fmt.Errorf("no command to stop in current path")
	}

	delete(
		agentConfig.LongRunningProcesses,
		env.ConfigLongRunningProcessWD(cmdWD),
	)

	return env.SaveConfigAsFile(
		config.ElevenAgentConfigFilePath,
		agentConfig,
	)
}

func tryToStartLongRunningProcess(
	cmdWD string,
	cmd string,
) (*proto.TryToStartLongRunningProcessReply, error) {

	grpcConn, err := grpc.Dial(
		config.GRPCServerURI,
		grpc.WithTransportCredentials(
			insecure.NewCredentials(),
		),
	)

	if err != nil {
		return nil, err
	}

	defer grpcConn.Close()

	agentClient := proto.NewAgentClient(grpcConn)

	initStream, err := agentClient.TryToStartLongRunningProcess(
		context.TODO(),
		&proto.TryToStartLongRunningProcessRequest{
			Cwd: cmdWD,
			Cmd: cmd,
		},
	)

	if err != nil {
		return nil, err
	}

	var reply *proto.TryToStartLongRunningProcessReply

	for {
		reply, err = initStream.Recv()

		if err != nil {
			return nil, err
		}

		if len(reply.Heartbeat) > 0 {
			continue
		}

		break
	}

	return reply, nil
}
