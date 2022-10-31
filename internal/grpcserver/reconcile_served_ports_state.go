package grpcserver

import (
	"github.com/eleven-sh/agent/config"
	"github.com/eleven-sh/agent/internal/caddy"
	"github.com/eleven-sh/agent/internal/env"
	"github.com/eleven-sh/agent/proto"
	"github.com/eleven-sh/eleven/entities"
)

func (*agentServer) ReconcileServedPortsState(
	req *proto.ReconcileServedPortsStateRequest,
	stream proto.Agent_ReconcileServedPortsStateServer,
) error {

	caddyConfig := caddy.CreateConfigFromServedPorts(req.ServedPorts)
	caddyAPI := caddy.NewAPI()

	err := caddyAPI.Load(caddyConfig)

	if err != nil {
		return err
	}

	agentConfig, err := env.LoadConfig(
		config.ElevenAgentConfigFilePath,
	)

	if err != nil {
		return err
	}

	agentConfig.ServedPorts = getConfigServedPortsFromProto(req.ServedPorts)

	return env.SaveConfigAsFile(
		config.ElevenAgentConfigFilePath,
		agentConfig,
	)
}

func getConfigServedPortsFromProto(
	servedPorts map[string]*proto.EnvServedPortBindings,
) env.ConfigServedPorts {

	configServedPorts := env.ConfigServedPorts{}

	for _, portBindings := range servedPorts {
		bindings := portBindings.Bindings

		for _, binding := range bindings {
			if binding.Type != string(entities.EnvServedPortBindingTypePort) {
				continue
			}

			configServedPorts[env.ConfigServedPort(binding.Value)] = true
		}
	}

	return configServedPorts
}
