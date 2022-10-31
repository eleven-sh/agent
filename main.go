package main

import (
	"log"
	"os"
	"time"

	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/internal/forever"
	"github.com/eleven-sh/agent/internal/grpcserver"
	"github.com/eleven-sh/agent/internal/sshserver"
	"github.com/eleven-sh/agent/internal/state"
)

type Command string

const (
	CommandForever Command = "forever"
)

var (
	SSHServerAuthorizedUsers = []sshserver.AuthorizedUser{
		{
			UserName:               config.ElevenUserName,
			AuthorizedKeysFilePath: config.ElevenUserAuthorizedSSHKeysFilePath,
		},
	}
)

func main() {
	// Given that logs are managed via journald
	// that always add the date and time before log lines
	// we remove it in the default Go logger
	log.SetFlags(0)

	// First element is binary path
	passedArgs := os.Args[1:]

	if len(passedArgs) > 0 {
		command := Command(passedArgs[0])

		if command == CommandForever {
			forever.Run(passedArgs[1:])
			return
		}

		log.Fatalf(
			"unsupported command: %s",
			command,
		)
	}

	go func() {
		log.Printf(
			"gRPC server listening at: %s",
			config.GRPCServerAddr,
		)

		err := grpcserver.ListenAndServe(
			config.GRPCServerAddrProtocol,
			config.GRPCServerAddr,
		)

		if err != nil {
			log.Fatalf("%v", err)
		}
	}()

	go func() {
		log.Printf(
			"Reconciling localhost proxies state...",
		)

		for {
			servedPorts := env.ConfigServedPorts{}

			agentConfig, err := env.LoadConfigIfExists(
				config.ElevenAgentConfigFilePath,
			)

			if err != nil {
				log.Fatalf("%v", err)
			}

			if agentConfig != nil {
				servedPorts = agentConfig.ServedPorts
			}

			err = state.ReconcileLocalhostProxies(servedPorts)

			if err != nil {
				log.Fatalf("%v", err)
			}

			time.Sleep(400 * time.Millisecond)
		}
	}()

	go func() {
		log.Printf(
			"Reconciling long running processes state...",
		)

		for {
			longRunningProcesses := env.ConfigLongRunningProcesses{}

			agentConfig, err := env.LoadConfigIfExists(
				config.ElevenAgentConfigFilePath,
			)

			if err != nil {
				log.Fatalf("%v", err)
			}

			if agentConfig != nil {
				longRunningProcesses = agentConfig.LongRunningProcesses
			}

			err = state.ReconcileLongRunningProcesses(longRunningProcesses)

			if err != nil {
				log.Fatalf("%v", err)
			}

			time.Sleep(400 * time.Millisecond)
		}
	}()

	sshServer, err := sshserver.NewServer(
		config.SSHServerHostKeyFilePath,
		SSHServerAuthorizedUsers,
		config.SSHServerListenAddr,
	)

	if err != nil {
		log.Fatalf("%v", err)
	}

	log.Printf(
		"SSH server listening at: %s",
		sshServer.Addr,
	)

	if err = sshServer.ListenAndServe(); err != nil {
		log.Fatalf("%v", err)
	}
}
