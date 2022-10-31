package grpcserver

import (
	_ "embed"
	"fmt"
	"net"
	"os"

	"github.com/eleven-sh/agent/proto"
	"google.golang.org/grpc"
)

type agentServer struct {
	proto.UnimplementedAgentServer
}

func ListenAndServe(serverAddrProtocol, serverAddr string) error {

	if serverAddrProtocol == "unix" {
		// Prevent "bind: address already in use" error
		if err := ensureOldServerUnixSocketRemoved(serverAddr); err != nil {
			return err
		}
	}

	tcpServer, err := net.Listen(serverAddrProtocol, serverAddr)

	if err != nil {
		return fmt.Errorf("failed to listen: %v", err)
	}

	if serverAddrProtocol == "unix" {
		if err := os.Chmod(serverAddr, 0600); err != nil {
			return fmt.Errorf("failed to set socket permissions: %v", err)
		}
	}

	grpcServer := grpc.NewServer()

	proto.RegisterAgentServer(grpcServer, &agentServer{})

	return grpcServer.Serve(tcpServer)
}

func ensureOldServerUnixSocketRemoved(socketPath string) error {
	return os.RemoveAll(socketPath)
}
