package grpcserver

import (
	"github.com/eleven-sh/agent/internal/caddy"
	"github.com/eleven-sh/agent/proto"
)

func (*agentServer) CheckDomainReachability(
	req *proto.CheckDomainReachabilityRequest,
	stream proto.Agent_CheckDomainReachabilityServer,
) error {

	caddyConfig := caddy.CreateConfigFromServedPorts(req.ServedPorts)

	caddy.UpdateConfigToCheckDomainReachability(
		caddyConfig,
		req.Domain,
		req.UniqueId,
	)

	caddyAPI := caddy.NewAPI()

	return caddyAPI.Load(caddyConfig)
}
